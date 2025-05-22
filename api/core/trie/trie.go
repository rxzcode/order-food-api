package trie

import (
	"strings"
)

const Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

var charToIndex [256]int

func init() {
	for i := range charToIndex {
		charToIndex[i] = -1
	}
	for i, ch := range Alphabet {
		charToIndex[ch] = i
	}
}

// Trie node
type TrieNode struct {
	children [36]*TrieNode
	isEnd    bool
}

type Trie struct {
	root *TrieNode
}

func NewTrie() *Trie {
	return &Trie{root: &TrieNode{}}
}

func (t *Trie) Insert(s string) bool {
	node := t.root
	s = strings.ToUpper(s)
	for _, ch := range s {
		idx := charToIndex[ch]
		if idx == -1 {
			return false
		}
		if node.children[idx] == nil {
			node.children[idx] = &TrieNode{}
		}
		node = node.children[idx]
	}
	node.isEnd = true
	return true
}

func (t *Trie) Search(s string) bool {
	node := t.root
	s = strings.ToUpper(s)
	for _, ch := range s {
		idx := charToIndex[ch]
		if idx == -1 {
			return false
		}
		if node.children[idx] == nil {
			return false
		}
		node = node.children[idx]
	}
	return node.isEnd
}
