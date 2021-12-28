package transaction

import (
	"bytes"
	"encoding/json"

	"filippo.io/edwards25519"
	"github.com/akamensky/base58"
	"github.com/timcki/learncoin/internal/crypto"
)

type curveValue interface {
	*edwards25519.Point | *edwards25519.Scalar
}

type PrivKey struct {
	a *edwards25519.Scalar
	b *edwards25519.Scalar
}

type PublicKey struct {
	A *edwards25519.Point
	B *edwards25519.Point
}

type Address struct {
	privKey PrivKey
	PubKey  PublicKey
}

type OneTimeAddress struct {
	P *edwards25519.Point
	R *edwards25519.Point
}

func (a OneTimeAddress) MarshalJSON() ([]byte, error) {
	result := make(map[string][]byte)
	result["P"] = a.P.Bytes()
	result["R"] = a.R.Bytes()
	return json.Marshal(&result)
}

func (a *OneTimeAddress) UnmarshalJSON([]byte) error {
	result := make(map[string][]byte)
	var err error
	a.P, err = edwards25519.NewIdentityPoint().SetBytes(result["P"])
	if err != nil {
		return err
	}
	a.R, err = edwards25519.NewIdentityPoint().SetBytes(result["R"])
	if err != nil {
		return err
	}
	return nil
}

func (pb PublicKey) ToHumanReadable(truncated bool) (string, error) {
	var buffer bytes.Buffer
	var addressType string
	var check []byte

	// Skip key A if address is truncated
	// Untruncated address begins with lrn1
	// Truncated with lrn0
	A := pb.A.Bytes()
	B := pb.B.Bytes()
	if !truncated {
		addressType = "lrn1"
		if _, err := buffer.Write(A); err != nil {
			return "", err
		}
	} else {
		addressType = "lrn0"
	}

	if _, err := buffer.Write(B); err != nil {
		return "", err
	}

	if !truncated {
		check = append(check, A...)
	}
	// Calculate the checksum (hash of both keys/key B if truncated)
	checksum, err := crypto.HashData(append(check, B...))
	if err != nil {
		return "", err
	}
	// writing first 8 bytes to compare
	if _, err := buffer.Write(checksum[:8]); err != nil {
		return "", err
	}

	return addressType + base58.Encode(buffer.Bytes()), nil
}

func NewPublicKeyFromHumanReadable(key string) {

}

// NewAddress generates a new address with entropy
func NewAddress() (Address, error) {
	var addr Address

	publicA, privateA, err := newKeypair()
	if err != nil {
		panic(err)
	}
	publicB, privateB, err := newKeypair()
	if err != nil {
		panic(err)
	}

	addr.privKey = PrivKey{a: privateA, b: privateB}
	addr.PubKey = PublicKey{A: publicA, B: publicB}

	return addr, nil
}

// CheckDestinationAddress checks if the destination address was generated from
// his public keyset:
// P' = Hs(aR)G + B
// if P' = P then Hs(aR)G + B = Hs(rA)G + B => arG = raG => true
func (a Address) CheckDestinationAddress(dest OneTimeAddress) bool {
	aR := new(edwards25519.Point).ScalarMult(a.privKey.a, dest.R)
	HsaR, err := hashPointToScalar(aR)
	// Ignore fault and return false
	if err != nil {
		return false
	}
	P := new(edwards25519.Point).Add(
		new(edwards25519.Point).ScalarBaseMult(HsaR),
		a.PubKey.B,
	)
	return P.Equal(dest.P) == 1
}

// Calculate P and R for destination address
// r - randomly generated values i [1, l-1]
// G - generator of ed25519 curve
// (A, B) -  public keys of recipient
// Calculates:
// P = Hs(rA)G + B
// R = rG
func (addr Address) NewDestinationAddress() (OneTimeAddress, error) {
	// Calculate random r and corresponding R
	// R = rG
	R, r, err := newKeypair()

	// Calculate rA
	rA := new(edwards25519.Point).ScalarMult(r, addr.PubKey.A)

	// Calculate Hs(rA)
	HsrA, err := hashPointToScalar(rA)
	if err != nil {
		return OneTimeAddress{}, err
	}

	// P = Hs(rA)G + B
	P := new(edwards25519.Point).Add(
		new(edwards25519.Point).ScalarBaseMult(HsrA),
		addr.PubKey.B,
	)

	return OneTimeAddress{P: P, R: R}, nil
}
