package mph

import (
	"sort"
)

const maxSeedAttempts = 1_000_000

type Table struct {
	keys       []uint64
	level0     []uint32
	level0Mask int
	level1     []uint32
	level1Mask int
	Extra      []uint64
}

// Build builds the MPH for string keys of max length 10, allowed chars 0-9 A-Z.
func Build(keys []string) *Table {
	encodedKeys := make([]uint64, len(keys))
	for i, k := range keys {
		encodedKeys[i] = encodeKey(k)
	}
	return buildUint64(encodedKeys)
}

func buildUint64(keys []uint64) *Table {
	if len(keys) == 0 {
		return &Table{}
	}

	level0Size := nextPow2(len(keys) / 8)
	level1Size := nextPow2(int(float64(len(keys)) * 0.9))

	level0 := make([]uint32, level0Size)
	level0Mask := level0Size - 1

	level1 := make([]uint32, level1Size)
	level1Mask := level1Size - 1

	sparseBuckets := make([][]int, level0Size)
	zeroSeed := murmurSeed(0)

	for i, k := range keys {
		n := int(zeroSeed.hashUint64(k)) & level0Mask
		sparseBuckets[n] = append(sparseBuckets[n], i)
	}

	var buckets []indexBucket
	for n, vals := range sparseBuckets {
		if len(vals) > 0 {
			buckets = append(buckets, indexBucket{n: n, vals: vals})
		}
	}
	sort.Sort(bySize(buckets))

	occ := make([]bool, level1Size)
	var tmpOcc []int
	var extra []uint64

	for _, bucket := range buckets {
		var seed murmurSeed
		found := false

		for attempt := 0; attempt < maxSeedAttempts; attempt++ {
			tmpOcc = tmpOcc[:0]
			collision := false

			for _, i := range bucket.vals {
				n := int(seed.hashUint64(keys[i])) & level1Mask
				if occ[n] {
					for _, o := range tmpOcc {
						occ[o] = false
					}
					seed++
					collision = true
					break
				}
				occ[n] = true
				tmpOcc = append(tmpOcc, n)
				level1[n] = uint32(i)
			}

			if !collision {
				level0[bucket.n] = uint32(seed)
				found = true
				break
			}
		}

		if !found {
			for _, i := range bucket.vals {
				extra = append(extra, keys[i])
			}
		}
	}

	return &Table{
		keys:       keys,
		level0:     level0,
		level0Mask: level0Mask,
		level1:     level1,
		level1Mask: level1Mask,
		Extra:      extra,
	}
}

// Lookup returns the index and if key exists.
func (t *Table) Lookup(key string) (n uint32, ok bool) {
	k := encodeKey(key)
	i0 := int(murmurSeed(0).hashUint64(k)) & t.level0Mask
	seed := t.level0[i0]
	i1 := int(murmurSeed(seed).hashUint64(k)) & t.level1Mask
	n = t.level1[i1]
	if n >= uint32(len(t.keys)) {
		return 0, false
	}
	if k == t.keys[int(n)] {
		return n, true
	}
	// fallback check in Extra (linear scan, can optimize with a map or sorted slice if needed)
	for _, x := range t.Extra {
		if x == k {
			return 0, true
		}
	}
	return 0, false
}

// encodeKey encodes up to 10 chars of 0-9A-Z into a uint64.
func encodeKey(key string) uint64 {
	var val uint64 = 0
	n := len(key)
	if n > 10 {
		n = 10
	}
	for i := 0; i < n; i++ {
		val <<= 6
		val |= charToVal(key[i])
	}
	// Shift leftover bits if key shorter than 10
	val <<= 6 * (10 - n)
	return val
}

// charToVal maps '0'-'9','A'-'Z' to 0-35.
func charToVal(c byte) uint64 {
	switch {
	case c >= '0' && c <= '9':
		return uint64(c - '0')
	case c >= 'A' && c <= 'Z':
		return uint64(c - 'A' + 10)
	default:
		return 0
	}
}

type indexBucket struct {
	n    int
	vals []int
}

type bySize []indexBucket

func (s bySize) Len() int           { return len(s) }
func (s bySize) Less(i, j int) bool { return len(s[i].vals) > len(s[j].vals) }
func (s bySize) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func nextPow2(n int) int {
	if n <= 0 {
		return 1
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	return n
}
