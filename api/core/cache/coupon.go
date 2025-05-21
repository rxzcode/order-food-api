package cache

import (
	"bytes"
	"log"
	"os"
	"sync"

	"github.com/cespare/xxhash/v2"
	"github.com/edsrzf/mmap-go"
)

type bitset uint8

type CouponCodeCache struct {
	mu    sync.RWMutex
	store map[uint64]bitset
}

func NewCouponCodeCache(filePaths []string) *CouponCodeCache {
	c := &CouponCodeCache{
		store: make(map[uint64]bitset, 100_000_000), // preallocate if needed
	}

	var wg sync.WaitGroup

	for idx, path := range filePaths {
		wg.Add(1)
		go func(p string, i int) {
			defer wg.Done()
			c.loadFile(p, i)
		}(path, idx)
	}

	go func() {
		wg.Wait()
		log.Println("All coupon codes loaded.")
	}()

	return c
}

func hashCode(code string) uint64 {
	return xxhash.Sum64String(code)
}

func (c *CouponCodeCache) loadFile(path string, fileIndex int) {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Failed to open %s: %v", path, err)
		return
	}
	defer file.Close()

	// Memory map the file
	mappedFile, err := mmap.Map(file, mmap.RDONLY, 0)
	if err != nil {
		log.Printf("Failed to mmap file %s: %v", path, err)
		return
	}
	defer mappedFile.Unmap()

	// Process lines from the memory-mapped file
	lines := bytes.Split(mappedFile, []byte("\n"))
	seen := make(map[uint64]bool)

	for i, line := range lines {
		if i%1_00_000 == 0 {
			log.Printf("LoadingCache [%s] %v", path, i)
		}
		code := string(line)
		hash := hashCode(code)
		if !seen[hash] {
			c.mu.Lock()
			c.store[hash] |= (1 << fileIndex)
			c.mu.Unlock()
			seen[hash] = true
		}
	}
	log.Printf("LoadingCache [%s] 100%%", path)
}

func (c *CouponCodeCache) Check(code string, minHits int) bool {
	hash := hashCode(code)

	c.mu.RLock()
	flags := c.store[hash]
	c.mu.RUnlock()

	count := 0
	for i := 0; i < 8; i++ {
		if (flags & (1 << i)) != 0 {
			count++
		}
	}
	return count >= minHits
}
