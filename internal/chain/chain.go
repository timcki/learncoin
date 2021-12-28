package chain

import (
	"sync"
	"time"

	"github.com/timcki/learncoin/internal/crypto"
	"github.com/timcki/learncoin/internal/transaction"
)

// Header is the header of a block
type Header struct {
	version      uint8
	previousHash crypto.Hash
	merkleRoot   crypto.Hash
	time         time.Time
}

type Block struct {
	header       Header
	transactions crypto.MerkleTree
}

type Chain struct {
	blocks []Block
	mu     sync.RWMutex
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
