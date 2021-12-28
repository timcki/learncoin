package transaction

import (
	"crypto/rand"

	"filippo.io/edwards25519"
	"github.com/timcki/learncoin/internal/crypto"
)

/*
func toBytes[T edwards25519.Point | edwards25519.Scalar](arr []*T) [][]byte {
	res := make([][]byte, len(arr))
	for i, val := range arr {
		res[i] = val.Bytes()
	}
	return res
}
*/

// ComputePrivateKey calculates x = Hs(aR) + b to the private key corresponding to the tx
// one time address, thus allowing to spend
func (addr Address) ComputePrivateKey(dest OneTimeAddress) (*edwards25519.Scalar, error) {
	aR := new(edwards25519.Point).ScalarMult(addr.privKey.a, dest.R)
	HsaR, err := hashPointToScalar(aR)
	if err != nil {
		return nil, err
	}
	x := new(edwards25519.Scalar).Add(HsaR, addr.privKey.b)
	return x, nil

}

// hashPointToScalar is used to compute Hs(xP) and convert it to a scalar
func hashPointToScalar(point *edwards25519.Point) (*edwards25519.Scalar, error) {
	rBytes, err := crypto.HashData(point.Bytes())
	if err != nil {
		return nil, err
	}
	return edwards25519.NewScalar().SetBytesWithClamping(rBytes)
}

// TODO: Implement
func HashPoint(p *edwards25519.Point) *edwards25519.Point {
	return p
}

func randomScalar() (priv *edwards25519.Scalar, err error) {
	seed := make([]byte, 64)
	rand.Read(seed)
	return new(edwards25519.Scalar).SetUniformBytes(seed)
}

func newKeypair() (public *edwards25519.Point, private *edwards25519.Scalar, err error) {
	private, err = randomScalar()
	public = new(edwards25519.Point).ScalarBaseMult(private)
	return
}
