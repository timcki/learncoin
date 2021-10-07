package crypto

import (
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

func (h Hash) String() string {
	return hex.EncodeToString(h)
}
