package chain

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"sync"
	"time"

	"github.com/akamensky/base58"
	"github.com/timcki/learncoin/internal/crypto"
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

type PrivKey struct {
	a ed25519.PrivateKey
	b ed25519.PrivateKey
}

type PublicKey struct {
	A ed25519.PublicKey
	B ed25519.PublicKey
}

func (pb PublicKey) ToHumanReadable() (string, error) {
	// Buffer that holds 0x96b9f4+<2 public keys>+checksum
	var buffer bytes.Buffer
	// Writing bytes to have every address begin by 38dF
	buffer.WriteRune('a')
	buffer.WriteRune('g')
	buffer.WriteRune('h')
	//buffer.WriteRune(0xf4)
	//buffer.WriteRune(0xb9)
	//buffer.WriteRune(0x96)
	if _, err := buffer.Write(pb.A); err != nil {
		return "", err
	}
	if _, err := buffer.Write(pb.B); err != nil {
		return "", err
	}
	checkA := pb.A
	//checksum is the hash of the two public keys
	checksum, err := crypto.HashData(append(checkA, pb.B...))
	if err != nil {
		return "", err
	}
	// writing first 8 bytes to compare
	if _, err := buffer.Write(checksum[:8]); err != nil {
		return "", err
	}

	return base58.Encode(buffer.Bytes()), nil
}

func NewPublicKeyFromHumanReadable(key string) {

}

type Address struct {
	privKey PrivKey
	PubKey  PublicKey
}

// NewAddress generates a new address with entropy
func NewAddress() (Address, error) {
	var addr Address
	puba, priva, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return addr, err
	}
	pubb, privb, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return addr, err
	}
	addr.privKey = PrivKey{a: priva, b: privb}
	addr.PubKey = PublicKey{A: puba, B: pubb}
	return addr, nil
}
