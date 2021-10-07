package crypto

import (
	"math"
)

type Node struct {
	leaf    bool
	Hash    Hash
	Content *Hashable
}

// TODO: To speedup lookup time create a hashtable with mapping
// tx_hash -> array_index (maybe overoptimization?)
type MerkleTree struct {
	nodes []Node
	depth uint
}

type EmptyLeaf int

func (e EmptyLeaf) Hash() (Hash, error) {
	h := make([]byte, 0)
	return h, nil
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
		childrenHash, err := HashData(append(tree[2*i+1].Hash, tree[2*i+2].Hash...))
		if err != nil {
			return nil, err
		}
		tree[i] = Node{leaf: false, Hash: childrenHash, Content: nil}
	}
	return &MerkleTree{tree, uint(baseIndex)}, nil
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

func (m *MerkleTree) GetNodes() []Node {
	return m.nodes
}

// TODO: Print the MerkleTree in a pretty form with hashes and values
func (m *MerkleTree) PresentTree() string {
	//result := ""
	return ""
}
