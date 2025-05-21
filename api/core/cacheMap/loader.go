package cacheMap

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	chunkSize   = 5_000_000
	workerCount = 32
)

type Loader struct {
	workerTables []map[string]map[string]struct{} // [worker]file -> set of codes
	lineChan     chan fileChunk
	wg           sync.WaitGroup
}

type fileChunk struct {
	fileName string
	lines    []string
}

func New() *Loader {
	workerTables := make([]map[string]map[string]struct{}, workerCount)
	for i := 0; i < workerCount; i++ {
		workerTables[i] = make(map[string]map[string]struct{})
	}

	return &Loader{
		workerTables: workerTables,
		lineChan:     make(chan fileChunk, workerCount*2),
	}
}

func (l *Loader) LoadFiles(files []string) error {
	for i := 0; i < workerCount; i++ {
		l.wg.Add(1)
		go l.worker(l.workerTables[i])
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

func (l *Loader) worker(localMap map[string]map[string]struct{}) {
	defer l.wg.Done()

	for job := range l.lineChan {
		codes, ok := localMap[job.fileName]
		if !ok {
			codes = make(map[string]struct{})
			localMap[job.fileName] = codes
		}

		for _, line := range job.lines {
			code := strings.TrimSpace(line)
			if code != "" {
				codes[code] = struct{}{}
			}
		}

		fmt.Printf("Loaded chunk for %s with %d codes\n", filepath.Base(job.fileName), len(job.lines))
	}
}

// AppearsInAtLeastN checks if the code appears in at least N different files.
func (l *Loader) AppearsInAtLeastN(code string, n int) bool {
	seenFiles := make(map[string]struct{})

	for _, workerMap := range l.workerTables {
		for file, codes := range workerMap {
			if _, alreadyCounted := seenFiles[file]; alreadyCounted {
				continue
			}
			if _, ok := codes[code]; ok {
				seenFiles[file] = struct{}{}
				if len(seenFiles) >= n {
					return true
				}
			}
		}
	}

	return false
}
