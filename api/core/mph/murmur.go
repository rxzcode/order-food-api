package mph

import (
	"encoding/binary"
	"reflect"
	"unsafe"
)

// This file contains an optimized murmur3 32-bit implementation tailored for
// our specific use case. See https://en.wikipedia.org/wiki/MurmurHash.

// A murmurSeed is the initial state of a Murmur3 hash.
type murmurSeed uint32

const (
	c1      = 0xcc9e2d51
	c2      = 0x1b873593
	r1Left  = 15
	r1Right = 32 - r1Left
	r2Left  = 13
	r2Right = 32 - r2Left
	m       = 5
	n       = 0xe6546b64
)

// murmur3 32-bit finalizer for uint64 key input.
func (m murmurSeed) hashUint64(key uint64) uint32 {
	// Use seed + key bytes as input to Murmur3-like hash

	// Convert uint64 key to 8 bytes
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], key)

	h := uint32(m)
	for _, b := range buf {
		h ^= uint32(b)
		h *= 0x5bd1e995
		h ^= h >> 15
	}
	// Final avalanche
	h ^= h >> 13
	h *= 0x5bd1e995
	h ^= h >> 15
	return h
}

// hash computes the 32-bit Murmur3 hash of s using ms as the seed.
func (ms murmurSeed) hash(s string) uint32 {
	h := uint32(ms)
	l := len(s)
	numBlocks := l / 4
	var blocks []uint32
	header := (*reflect.SliceHeader)(unsafe.Pointer(&blocks))
	header.Data = (*reflect.StringHeader)(unsafe.Pointer(&s)).Data
	header.Len = numBlocks
	header.Cap = numBlocks
	for _, k := range blocks {
		k *= c1
		k = (k << r1Left) | (k >> r1Right)
		k *= c2
		h ^= k
		h = (h << r2Left) | (h >> r2Right)
		h = h*m + n
	}

	var k uint32
	ntail := l & 3
	itail := l - ntail
	switch ntail {
	case 3:
		k ^= uint32(s[itail+2]) << 16
		fallthrough
	case 2:
		k ^= uint32(s[itail+1]) << 8
		fallthrough
	case 1:
		k ^= uint32(s[itail])
		k *= c1
		k = (k << r1Left) | (k >> r1Right)
		k *= c2
		h ^= k
	}

	h ^= uint32(l)
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16
	return h
}
