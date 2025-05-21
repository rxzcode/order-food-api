package search

import (
	"context"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

type Loader struct {
	filePaths []string
}

func New() *Loader {
	return &Loader{}
}

func (l *Loader) LoadFiles(paths []string) error {
	l.filePaths = paths
	return nil
}

func (l *Loader) AppearsInAtLeastN(str string, n int) bool {
	numWorkers := runtime.NumCPU()
	var wg sync.WaitGroup
	matchCh := make(chan struct{}, len(l.filePaths))
	stopCh := make(chan struct{})
	once := sync.Once{}

	fileCh := make(chan string)

	// Worker goroutines using rg
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range fileCh {
				select {
				case <-stopCh:
					return
				default:
					if searchWithRipgrep(filePath, str) {
						select {
						case matchCh <- struct{}{}:
						case <-stopCh:
						}
					}
				}
			}
		}()
	}

	// Send file paths to workers
	go func() {
		for _, path := range l.filePaths {
			select {
			case <-stopCh:
				break
			default:
				fileCh <- path
			}
		}
		close(fileCh)
	}()

	// Close matchCh once all workers are done
	go func() {
		wg.Wait()
		close(matchCh)
	}()

	// Count matches
	count := 0
	for range matchCh {
		count++
		if count >= n {
			once.Do(func() {
				close(stopCh)
			})
			return true
		}
	}
	once.Do(func() {
		close(stopCh)
	})
	return false
}

func searchWithRipgrep(filePath, pattern string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "rg", "--search-zip", pattern, filePath)
	err := cmd.Run()

	return err == nil // true if match found
}
