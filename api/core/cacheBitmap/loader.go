package cacheMap

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/RoaringBitmap/roaring/v2"
	"github.com/cespare/xxhash/v2"
)

const (
	chunkSize   = 5_000_000
	workerCount = 32
)

type fileChunk struct {
	fileName string
	lines    []string
}

type Loader struct {
	lineChan     chan fileChunk
	wg           sync.WaitGroup
	fileBitmaps  map[string]*roaring.Bitmap
	fileBitmapsM sync.Mutex
}

// New returns a new loader
func New() *Loader {
	return &Loader{
		lineChan:    make(chan fileChunk, workerCount*2),
		fileBitmaps: make(map[string]*roaring.Bitmap),
	}
}

// LoadFiles reads all files in chunks using gzip scanner and feeds to workers
func (l *Loader) LoadFiles(files []string) error {
	for i := 0; i < workerCount; i++ {
		l.wg.Add(1)
		go l.worker()
	}

	for _, file := range files {
		if err := l.loadFileChunks(file); err != nil {
			return fmt.Errorf("load %s: %w", file, err)
		}
	}

	close(l.lineChan)
	l.wg.Wait()
	return nil
}

// Reads a .gz file in chunks and sends them to lineChan
func (l *Loader) loadFileChunks(path string) error {
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
			l.sendChunk(path, chunk)
			chunk = chunk[:0]
		}
	}
	if len(chunk) > 0 {
		l.sendChunk(path, chunk)
	}

	return scanner.Err()
}

func (l *Loader) sendChunk(fileName string, lines []string) {
	copyLines := make([]string, len(lines))
	copy(copyLines, lines)
	l.lineChan <- fileChunk{fileName: fileName, lines: copyLines}
}

// Each worker consumes chunks and builds partial bitmap, merges into file-level bitmap
func (l *Loader) worker() {
	defer l.wg.Done()

	for chunk := range l.lineChan {
		bm := roaring.New()
		for _, line := range chunk.lines {
			code := strings.TrimSpace(line)
			if code != "" {
				bm.Add(hashToUint32(code))
			}
		}

		l.fileBitmapsM.Lock()
		existing, ok := l.fileBitmaps[chunk.fileName]
		if ok {
			existing.Or(bm)
		} else {
			l.fileBitmaps[chunk.fileName] = bm
		}
		l.fileBitmapsM.Unlock()

		fmt.Printf("Processed chunk for %s with %d codes\n", filepath.Base(chunk.fileName), len(chunk.lines))
	}
}

// AppearsInAtLeastN returns true if code appears in at least n files
func (l *Loader) AppearsInAtLeastN(code string, n int) bool {
	id := hashToUint32(code)

	l.fileBitmapsM.Lock()
	defer l.fileBitmapsM.Unlock()

	found := 0
	for _, bm := range l.fileBitmaps {
		if bm.Contains(id) {
			found++
			if found >= n {
				return true
			}
		}
	}
	return false
}

func hashToUint32(s string) uint32 {
	return uint32(xxhash.Sum64String(s))
}
