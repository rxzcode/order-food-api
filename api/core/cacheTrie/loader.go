package cacheTrie

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"order-food-api/core/trie"
	"os"
	"path/filepath"
	"sync"
)

const (
	chunkSize   = 500_000
	workerCount = 16
)

type Loader struct {
	mu         sync.Mutex
	fileTables map[string][]*trie.Trie // file -> tables
	wg         sync.WaitGroup
	lineChan   chan fileChunk
}

type fileChunk struct {
	fileName string
	lines    []string
}

func New() *Loader {
	return &Loader{
		fileTables: make(map[string][]*trie.Trie),
		lineChan:   make(chan fileChunk, workerCount*2),
	}
}

func (l *Loader) LoadFiles(files []string) error {
	for i := 0; i < workerCount; i++ {
		l.wg.Add(1)
		go l.worker()
	}

	for _, file := range files {
		if err := l.loadFile(file); err != nil {
			return fmt.Errorf("error loading %s: %w", file, err)
		}
	}

	close(l.lineChan)
	l.wg.Wait()
	return nil
}

func (l *Loader) loadFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	r, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer r.Close()

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	var chunk []string
	for scanner.Scan() {
		chunk = append(chunk, scanner.Text())
		if len(chunk) >= chunkSize {
			copyChunk := make([]string, len(chunk))
			copy(copyChunk, chunk)
			l.lineChan <- fileChunk{fileName: path, lines: copyChunk}
			chunk = chunk[:0]
		}
	}
	if len(chunk) > 0 {
		copyChunk := make([]string, len(chunk))
		copy(copyChunk, chunk)
		l.lineChan <- fileChunk{fileName: path, lines: copyChunk}
	}
	return scanner.Err()
}

func (l *Loader) worker() {
	defer l.wg.Done()
	for job := range l.lineChan {
		table := trie.NewTrie()
		for _, line := range job.lines {
			table.Insert(line)
		}
		l.mu.Lock()
		l.fileTables[job.fileName] = append(l.fileTables[job.fileName], table)
		l.mu.Unlock()

		fmt.Printf("Built table for %s with %d entries\n",
			filepath.Base(job.fileName), len(job.lines))
	}
}

func (l *Loader) AppearsInAtLeastN(code string, n int) bool {
	count := 0

	l.mu.Lock()
	defer l.mu.Unlock()

	for _, tables := range l.fileTables {
		found := false

		for _, t := range tables {
			if ok := t.Search(code); ok {
				found = true
				break
			}
			if found {
				break
			}
		}

		if found {
			count++
			if count >= n {
				return true
			}
		}
	}

	return false
}
