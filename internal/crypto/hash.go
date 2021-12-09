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
type FixedHash [16]byte

func (h Hash) String() string {
	return hex.EncodeToString(h)
}

func (h Hash) ToFixedHash() FixedHash {
	var fixed FixedHash
	copy(fixed[:], h[:])
	return fixed
}
