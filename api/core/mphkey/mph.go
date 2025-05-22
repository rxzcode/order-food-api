package mphkey

const maxSeedAttempts = 1_000_000

type Table struct {
	keys  []uint64
	Extra []uint64
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

	return &Table{
		keys:  keys,
		Extra: []uint64{},
	}
}

func indexOf(slice []uint64, val uint64) int {
	for i, v := range slice {
		if v == val {
			return i
		}
	}
	return -1 // not found
}

// Lookup returns the index and if key exists.
func (t *Table) Lookup(key string) (n uint32, ok bool) {
	k := encodeKey(key)
	if indexOf(t.keys, k) != -1 {
		return n, true
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

type murmurSeed uint32

func (m murmurSeed) hashUint64(key uint64) uint32 {
	const c1 uint64 = 0x87c37b91114253d5
	const c2 uint64 = 0x4cf5ad432745937f

	k := key
	k *= c1
	k = (k << 31) | (k >> (64 - 31))
	k *= c2

	h1 := uint64(m) ^ k
	h1 = (h1 << 27) | (h1 >> (64 - 27))
	h1 = h1*5 + 0x52dce729

	h1 ^= 8 // number of bytes processed
	h1 ^= h1 >> 33
	h1 *= 0xff51afd7ed558ccd
	h1 ^= h1 >> 33
	h1 *= 0xc4ceb9fe1a85ec53
	h1 ^= h1 >> 33

	return uint32(h1)
}
