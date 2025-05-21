package cacheBloomFilter

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bits-and-blooms/bloom/v3"
)

const (
	chunkSize          = 5_000_000
	workerCount        = 32
	bloomExpectedItems = 10_000_000
	bloomFalsePositive = 0.00001
)

type Loader struct {
	workerTables []map[string]*bloom.BloomFilter
	lineChan     chan fileChunk
	wg           sync.WaitGroup
}

type fileChunk struct {
	fileName string
	lines    []string
}

func NewCache() *Loader {
	workerTables := make([]map[string]*bloom.BloomFilter, workerCount)
	for i := 0; i < workerCount; i++ {
		workerTables[i] = make(map[string]*bloom.BloomFilter)
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

func (l *Loader) worker(localMap map[string]*bloom.BloomFilter) {
	defer l.wg.Done()

	for job := range l.lineChan {
		filter, ok := localMap[job.fileName]
		if !ok {
			filter = bloom.NewWithEstimates(bloomExpectedItems, bloomFalsePositive)
			localMap[job.fileName] = filter
		}

		for _, line := range job.lines {
			code := strings.TrimSpace(line)
			if code != "" {
				filter.AddString(code)
			}
		}

		fmt.Printf("Loaded chunk for %s with %d codes\n", filepath.Base(job.fileName), len(job.lines))
	}
}

func (l *Loader) AppearsInAtLeastN(code string, n int) bool {
	seen := 0
	checked := make(map[string]struct{})

	for _, workerMap := range l.workerTables {
		for file, filter := range workerMap {
			if _, ok := checked[file]; ok {
				continue
			}
			checked[file] = struct{}{}
			if filter.TestString(code) {
				seen++
				if seen >= n {
					return true
				}
			}
		}
	}

	return false
}
