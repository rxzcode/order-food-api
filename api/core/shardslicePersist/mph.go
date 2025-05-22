package shardslice

import (
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"os"
	"path/filepath"
)

const maxSeedAttempts = 1_000_000
const tableDir = "./glob/tables"

type Table struct {
	Hash  string
	Keys  []uint64
	Extra []uint64
}

// Build encodes keys, builds table, saves to disk, and clears memory.
func Build(keys []string) *Table {
	encodedKeys := make([]uint64, len(keys))
	for i, k := range keys {
		encodedKeys[i] = encodeKey(k)
	}
	t := buildUint64(encodedKeys)
	t.Hash = hashKeys(keys)

	if err := saveTable(t); err != nil {
		panic(err) // or handle gracefully
	}

	t.Keys = nil // Release memory after saving
	return t
}

// buildUint64 constructs a Table from encoded keys.
func buildUint64(keys []uint64) *Table {
	if len(keys) == 0 {
		return &Table{}
	}
	return &Table{
		Keys:  keys,
		Extra: []uint64{},
	}
}

// Lookup loads the keys from file if needed, checks for key, and releases memory.
func (t *Table) Lookup(key string) (n uint32, ok bool) {
	if len(t.Keys) == 0 {
		if err := t.loadFromFile(); err != nil {
			return 0, false
		}
	}

	k := encodeKey(key)
	found := indexOf(t.Keys, k) != -1

	t.Keys = nil // Release after lookup
	return n, found
}

// encodeKey encodes a string into a compact uint64.
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

// hashKeys creates a hash from the input key list (used for filename).
func hashKeys(keys []string) string {
	h := sha256.New()
	for _, k := range keys {
		h.Write([]byte(k))
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// indexOf finds the position of a value in a slice, or returns -1 if not found.
func indexOf(slice []uint64, val uint64) int {
	for i, v := range slice {
		if v == val {
			return i
		}
	}
	return -1
}

// saveTable saves the Keys slice to a file named by hash.
func saveTable(t *Table) error {
	if err := os.MkdirAll(tableDir, 0755); err != nil {
		return err
	}
	path := filepath.Join(tableDir, t.Hash+".gob")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	return enc.Encode(t.Keys)
}

// loadFromFile loads the Keys slice from the file by hash.
func (t *Table) loadFromFile() error {
	path := filepath.Join(tableDir, t.Hash+".gob")
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var keys []uint64
	dec := gob.NewDecoder(f)
	if err := dec.Decode(&keys); err != nil {
		return err
	}

	t.Keys = keys
	return nil
}
