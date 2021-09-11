package main

import (
	"crypto/sha512"
	"encoding/hex"
)

// Hashable guarantees that a given type implements
// the Hash function
type Hashable interface {
	Hash() (Hash, error)
}

// Hash is a convenience type used for representing a
// hash digest
type Hash []byte

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
