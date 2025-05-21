package mph

import (
	"sort"
)

const maxSeedAttempts = 1_000_000 // Prevent infinite seed search

type Table struct {
	keys       []string
	level0     []uint32
	level0Mask int
	level1     []uint32
	level1Mask int
	Extra      []string
}

func Build(keys []string) *Table {
	if len(keys) == 0 {
		return &Table{}
	}

	level0 := make([]uint32, nextPow2(len(keys)/4))
	level0Mask := len(level0) - 1

	level1 := make([]uint32, nextPow2(len(keys)))
	level1Mask := len(level1) - 1

	sparseBuckets := make([][]int, len(level0))
	zeroSeed := murmurSeed(0)

	// Assign each key to a bucket in level0
	for i, s := range keys {
		n := int(zeroSeed.hash(s)) & level0Mask
		sparseBuckets[n] = append(sparseBuckets[n], i)
	}

	// Sort buckets by decreasing size (largest first)
	var buckets []indexBucket
	for n, vals := range sparseBuckets {
		if len(vals) > 0 {
			buckets = append(buckets, indexBucket{n: n, vals: vals})
		}
	}
	sort.Sort(bySize(buckets))

	// Occupancy tracker for level1
	occ := make([]bool, len(level1))
	var tmpOcc []int
	var extra []string

	for _, bucket := range buckets {
		var seed murmurSeed
		found := false

		for attempt := 0; attempt < maxSeedAttempts; attempt++ {
			tmpOcc = tmpOcc[:0]
			collision := false

			for _, i := range bucket.vals {
				n := int(seed.hash(keys[i])) & level1Mask
				if occ[n] {
					// Reset tmpOcc
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
			// log.Printf("Bucket %d failed, falling back to extra store (size %d)", bucket.n, len(bucket.vals))
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

func nextPow2(n int) int {
	for i := 1; ; i *= 2 {
		if i >= n {
			return i
		}
	}
}

func (t *Table) Lookup(s string) (n uint32, ok bool) {
	i0 := int(murmurSeed(0).hash(s)) & t.level0Mask
	seed := t.level0[i0]
	i1 := int(murmurSeed(seed).hash(s)) & t.level1Mask
	n = t.level1[i1]
	return n, s == t.keys[int(n)]
}

type indexBucket struct {
	n    int
	vals []int
}

type bySize []indexBucket

func (s bySize) Len() int           { return len(s) }
func (s bySize) Less(i, j int) bool { return len(s[i].vals) > len(s[j].vals) }
func (s bySize) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
