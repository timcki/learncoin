package main

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"math"
	"time"
)

// Hashable guarantees that a given type implements
// the Hash function
type Hashable interface {
	Hash() (Hash, error)
}

// Hash is a convenience type used for representing a
// hash digest
type Hash []byte

// Header is the header of a block
type Header struct {
	version       uint8
	previous_hash Hash
	merkle_root   Hash
	time          time.Time
}

func hash(data []byte) (Hash, error) {
	sum := sha512.New()
	if _, err := sum.Write(data); err != nil {
		return nil, err
	}
	return sum.Sum(nil), nil
}

func (h Hash) String() string {
	return hex.EncodeToString(h)
}

type Node struct {
	leaf    bool
	Hash    Hash
	Content *Hashable
}

type EmptyLeaf int

func (e EmptyLeaf) Hash() (Hash, error) {
	h := make([]byte, 0)
	return h, nil
}

// TODO: To speedup lookup time create a hashtable with mapping
// tx_hash -> array_index (maybe overoptimization?)
type MerkleTree struct {
	nodes []Node
}

// NewMerkleTree constructs a binary tree in which
// the leaves are transactions and parents are subsequent
// hashes. Since we know the depth at which the children will start
func NewMerkleTree(elements []Hashable) (*MerkleTree, error) {
	var pow float64
	elements, pow = fillElements(elements)
	// Tree size is sum(2^n for n from pow to 1)
	// so also 2^(n+1) - 1
	size := int(math.Pow(2, pow+1) - 1)
	tree := make([]Node, size)

	// Calculate the base index for level of leaves
	baseIndex := int(math.Pow(2, pow)) - 1
	// Fill all of the leaves with txs
	for i, element := range elements {
		hash, err := element.Hash()
		if err != nil {
			return nil, err
		}
		tree[baseIndex+i] = Node{leaf: true, Hash: hash, Content: &element}
	}
	// Fill all parent nodes until root
	for i := baseIndex - 1; i >= 0; i-- {
		childrenHash, err := hash(append(tree[2*i+1].Hash, tree[2*i+2].Hash...))
		if err != nil {
			return nil, err
		}
		tree[i] = Node{leaf: false, Hash: childrenHash, Content: nil}
	}
	return &MerkleTree{tree}, nil
}

// fillElements rounds the number of leafs to the nearest
// power of 2 greater than the length
func fillElements(el []Hashable) ([]Hashable, float64) {
	l := len(el)
	// Calculate the nearest power of two greater than the len of our tx list
	pow := math.Ceil(math.Log2(float64(l)))
	for i := 0; i < int(math.Pow(2, pow))-l; i++ {
		el = append(el, EmptyLeaf(0))
	}
	return el, pow
}

func (m *MerkleTree) RootHash() Hash {
	return m.nodes[0].Hash
}

func (m *MerkleTree) PrintTree() string {
	result := ""

}

type Block struct {
	header       Header
	transactions MerkleTree
}

type Blockchain struct {
	blocks []Block
}

func main() {
	txs := []Hashable{&Transaction{12}, &Transaction{24}, &Transaction{36}}
	for _, tx := range txs {
		hash, _ := tx.Hash()
		fmt.Println(hash.String())
	}
	tree, err := NewMerkleTree(txs)
	if err != nil {
		panic(err)
	}
	for i, el := range tree.nodes {
		fmt.Println(i, el)

	}
}
