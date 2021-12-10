package transaction

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/timcki/learncoin/internal/chain"
	"github.com/timcki/learncoin/internal/crypto"
)

// Address is a mock structure for representing addresses
// TODO: Implement valid addresses

// Utxo represents an unspent transaction output that's used in actual transactions to move value between addresses

type Utxo struct {
	Amount float32
}

type Transaction struct {
	TxINs  []*Utxo
	TxOUTs []*Utxo
	From   chain.Address
	To     chain.Address
}

// CheckValidity performs checks making sure that the txn is valid
func (t Transaction) CheckValidity() bool {
	for _, txin := range t.TxINs {
		fmt.Println(txin)
	}
	return false
}

// Used for testing purposes
//func randomTransaction() *Transaction {
//sender := Address{rand.Intn(16384), rand.Intn(16384)}
//receiver := Address{rand.Intn(16384), rand.Intn(16384)}
//return &Transaction{sender, receiver, rand.Intn(1000)}
//}

//func NewTransaction(sender, receiver Address, amount int) Transaction {
//return Transaction{Sender: sender, Receiver: receiver, Amount: amount}
//}

// I'm choosing the default Go byte encoder (gob). It reduces interoperability
// but it's not important since this is a reference design.
// In case it's a problem it's also abstrated away so easy to change
func (t *Transaction) toBytes() []byte {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	encoder.Encode(t)
	return buffer.Bytes()
}

func (t *Transaction) Hash() (crypto.Hash, error) {
	if h, err := crypto.HashData(t.toBytes()); err != nil {
		return nil, err
	} else {
		return h, nil
	}
}
