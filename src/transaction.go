package main

import "encoding/binary"

type Transaction struct {
	Amount int
}

func (t *Transaction) toBytes() []byte {
	r := make([]byte, 4)
	binary.LittleEndian.PutUint32(r, uint32(t.Amount))
	return r
}

func (t *Transaction) Hash() (Hash, error) {
	if h, err := hash(t.toBytes()); err != nil {
		return nil, err
	} else {
		return h, nil
	}
}
