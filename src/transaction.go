package main

import (
	"bytes"
	"encoding/gob"
	"math/rand"
)

// Address is a mock structure for representing addresses
// TODO: Implement valid addresses
type Address struct {
	PrivKey int
	PubKey  int
}

type Transaction struct {
	Sender   Address
	Receiver Address
	Amount   int
}

// Used for testing purposes
func randomTransaction() *Transaction {
	sender := Address{rand.Intn(16384), rand.Intn(16384)}
	receiver := Address{rand.Intn(16384), rand.Intn(16384)}
	return &Transaction{sender, receiver, rand.Intn(1000)}
}

func NewTransaction(sender, receiver Address, amount int) Transaction {
	return Transaction{Sender: sender, Receiver: receiver, Amount: amount}
}

// I'm choosing the default Go byte encoder (gob). It reduces interoperability
// but it's not important since this is a reference design.
// In case it's a problem it's also abstrated away so easy to change
func (t *Transaction) toBytes() []byte {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	encoder.Encode(t)
	return buffer.Bytes()
}

func (t *Transaction) Hash() (Hash, error) {
	if h, err := hash(t.toBytes()); err != nil {
		return nil, err
	} else {
		return h, nil
	}
}
