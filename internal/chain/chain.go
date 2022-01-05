package chain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/timcki/learncoin/internal/crypto"
	"github.com/timcki/learncoin/internal/transaction"
)

// Header is the header of a block
type Header struct {
	Version      uint8       `json:"version"`
	PreviousHash crypto.Hash `json:"previous_hash"`
	MerkleRoot   crypto.Hash `json:"merkle_root"`
	hash         crypto.Hash `json:"hash"`
	Time         time.Time   `json:"time"`
}

// Block is a container for groups of transactions. It
type Block struct {
	Header       Header
	Transactions crypto.MerkleTree
}

// Chain is the abstraction of a blockchain, which means
// * blocks represents a slice of blocks that expands
// * a mutex that allows multi-threaded reads/writes
type Chain struct {
	blocks []*Block
	mu     sync.RWMutex
}

func (h Header) PrettyPrint() string {
	var str strings.Builder
	str.WriteString("{\n")
	str.WriteString(fmt.Sprintf("  previous_hash: %v,\n", h.PreviousHash))
	str.WriteString(fmt.Sprintf("  merkle_root:   %v,\n", h.MerkleRoot))
	str.WriteString(fmt.Sprintf("  hash:          %v,\n", h.hash))
	str.WriteString(fmt.Sprintf("  time:          %v\n}", h.Time))

	return str.String()

}

// Bytes returns a byte array representation the Header
func (h Header) Bytes() []byte {
	var buf bytes.Buffer

	buf.WriteByte(byte(h.Version))
	buf.Write(h.MerkleRoot)
	buf.WriteString(strconv.Itoa(int(h.Time.Unix())))

	return buf.Bytes()
}

// Hash computes the Hash of Header
func (h Header) Hash() (crypto.Hash, error) {

	if h, err := crypto.HashData(h.Bytes()); err != nil {
		return nil, err
	} else {
		return h, nil
	}

}

func (b *Block) PrettyPrint() string {
	res, _ := json.MarshalIndent(b, "", "  ")
	return string(res)
}

func (c *Chain) Length() int {
	return len(c.blocks)
}

func NewBlock(txns []crypto.Hashable) *Block {
	merkleTree, _ := crypto.NewMerkleTree(txns)
	header := Header{
		Version:    1,
		MerkleRoot: merkleTree.RootHash(),
		Time:       time.Now(),
	}

	block := Block{
		Header:       header,
		Transactions: *merkleTree,
	}
	block.Header.hash, _ = block.Header.Hash()
	return &block
}

func (b *Block) SetPreviousHash(h crypto.Hash) {
	b.Header.PreviousHash = h
}

func (c *Chain) AddBlock(block *Block) {
	c.mu.Lock()
	prev_block_hash, err := c.blocks[len(c.blocks)-1].Header.Hash()
	if err != nil {
		panic(err)
	}
	block.SetPreviousHash(prev_block_hash)
	c.blocks = append(c.blocks, block)
	c.mu.Unlock()
}

func NewChain() *Chain {
	genesis := Block{
		Header: Header{
			Version:      0,
			PreviousHash: []byte{0},
			MerkleRoot:   []byte{0},
			hash:         []byte{0},
			Time:         time.Time{},
		},
		Transactions: crypto.MerkleTree{},
	}
	return &Chain{
		blocks: []*Block{&genesis},
		mu:     sync.RWMutex{},
	}

}

type UtxoSet interface {
	Add(transaction.Utxo) error
	UtxoIn(transaction.Utxo) bool
	Remove(transaction.Utxo) error
	GetUtxos() []*transaction.Utxo
	Get(crypto.FixedHash) *transaction.Utxo
}

type utxoSet struct {
	set map[crypto.FixedHash]*transaction.Utxo
}

func NewUtxoSet() UtxoSet {
	return &utxoSet{
		set: make(map[crypto.FixedHash]*transaction.Utxo),
	}
}

func (u *utxoSet) Add(utxo transaction.Utxo) error {
	h, err := utxo.Hash()
	if err != nil {
		return err
	}

	u.set[h.ToFixedHash()] = &utxo
	return nil
}

func (u *utxoSet) UtxoIn(utxo transaction.Utxo) bool {
	h, err := utxo.Hash()
	if err != nil {
		return false
	}
	_, k := u.set[h.ToFixedHash()]
	return k
}

func (u *utxoSet) Remove(utxo transaction.Utxo) error {
	h, err := utxo.Hash()
	if err != nil {
		return err
	}
	delete(u.set, h.ToFixedHash())
	return nil
}

func (u *utxoSet) GetUtxos() []*transaction.Utxo {
	ret := make([]*transaction.Utxo, 0)
	for _, v := range u.set {
		ret = append(ret, v)
	}
	return ret
}

func (u *utxoSet) Get(h crypto.FixedHash) *transaction.Utxo {
	return u.set[h]
}
