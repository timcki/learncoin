package transaction

import (
	"bytes"
	"fmt"

	"filippo.io/edwards25519"
	"github.com/timcki/learncoin/internal/crypto"
	"github.com/timcki/learncoin/internal/utility"
)

type RingSignature struct {
	// Message for which the ring sig is computed
	Utxos []Utxo
	// Key image in byte representation
	Image []byte
	// Challenges in byte representation
	C [][]byte
	// Responses in byte representation
	R [][]byte
}

func KeyImage(addr Address, dest OneTimeAddress) (x *edwards25519.Scalar, img *edwards25519.Point) {
	// Compute the txn private key x to use in the key image computation
	var err error
	if x, err = addr.ComputePrivateKey(dest); err != nil {
		panic(err)
	}
	img = new(edwards25519.Point).ScalarMult(x, HashPoint(dest.P))
	return
}

// ComputeChallenge is used to compute c.
// c is used in computing the challenge for the real txn
// in the ring set
// TODO: single UTXO or entire Transaction ?
func ComputeChallenge(messages []Utxo, L []*edwards25519.Point, R []*edwards25519.Point) (*edwards25519.Scalar, error) {
	var buffer bytes.Buffer

	// Write message (utxo in bytes representation) to buffer
	for _, message := range messages {
		buffer.Write(message.Bytes())
	}

	// Write consecutive L_i to buffer
	for _, Li := range L {
		buffer.Write(Li.Bytes())
	}
	// Write consecutive R_i to buffer
	for _, Ri := range R {
		buffer.Write(Ri.Bytes())
	}

	// Calculate c = Hs(m, L_0, .., L_n, R_0, .., R_n)
	if hash, err := crypto.HashData(buffer.Bytes()); err != nil {
		return nil, err
	} else {
		return edwards25519.NewScalar().SetBytesWithClamping(hash)
	}
}

// NewRingSignature returns a ring sig for given utxo and array of decoys with same value
// To create a ring signature we need the priv key of our destination address
// and n pubkeys from other txns with the same value as ours
func (a Address) NewRingSignature(realTxn Utxo, decoyTxns []Utxo) RingSignature {

	//d.P.MultiScalarMult()

	x, I := KeyImage(a, realTxn.Keypair)
	truePos, txns := utility.ShuffleAndAdd(realTxn, decoyTxns)
	// Compute len of our ringset n
	n := len(txns)

	// We'll use q_i in every single operation
	q := make([]*edwards25519.Scalar, n)
	for i := 0; i < n; i++ {
		q[i], _ = randomScalar()
	}

	// No need for w_i where i == s; s -> our priv key
	w := make([]*edwards25519.Scalar, n)
	for i := 0; i < n; i++ {
		// Zero our the co-efficient for our key
		if i == truePos {
			w[i] = edwards25519.NewScalar()
		} else {
			w[i], _ = randomScalar()
		}
	}

	L := make([]*edwards25519.Point, n)
	for i := 0; i < n; i++ {
		// TODO: Check if multiplying by zero actually gives us the right result
		L[i] = new(edwards25519.Point).Add(
			new(edwards25519.Point).ScalarBaseMult(q[i]),
			new(edwards25519.Point).ScalarMult(w[i], txns[i].Keypair.P),
		)
	}
	R := make([]*edwards25519.Point, n)
	for i := 0; i < n; i++ {
		R[i] = new(edwards25519.Point).Add(
			new(edwards25519.Point).ScalarMult(q[i], HashPoint(txns[i].Keypair.P)),
			new(edwards25519.Point).ScalarMult(w[i], I),
		)
	}

	// Compute the challenge
	c, err := ComputeChallenge(txns, L, R)
	if err != nil {
		panic(err)
	}

	// Compute the components for our response
	l := make([]*edwards25519.Scalar, n)
	// Prepare component to update sum as we compute consecutive values
	sumCi := edwards25519.NewScalar()

	for i := 0; i < n; i++ {
		// Skip computing for true utxo until we have every other
		if i != truePos {
			l[i] = w[i]
			sumCi = edwards25519.NewScalar().Add(sumCi, w[i])
		}
	}
	// Compute value for true utxo. Since the addition is modular it's unrecoverable
	l[truePos] = edwards25519.NewScalar().Subtract(c, sumCi)

	r := make([]*edwards25519.Scalar, n)
	for i := 0; i < n; i++ {
		if i == truePos {
			r[i] = edwards25519.NewScalar().Subtract(
				q[i],
				edwards25519.NewScalar().Multiply(l[i], x),
			)
		} else {
			r[i] = q[i]
		}
	}
	return RingSignature{
		Utxos: txns,
		Image: I.Bytes(),
		// Convert scalar/point values to []byte
		C: func() [][]byte {
			ret := make([][]byte, n)
			for i, val := range l {
				ret[i] = val.Bytes()
			}
			return ret
		}(),
		R: func() [][]byte {
			ret := make([][]byte, n)
			for i, val := range r {
				ret[i] = val.Bytes()
			}
			return ret
		}(),
	}
}

// Convert C scalar values to proper scalar type from byte representation
func (ringSig RingSignature) CRToScalars() (C, R []*edwards25519.Scalar, err error) {
	for _, val := range ringSig.C {
		var scalar *edwards25519.Scalar
		scalar, err = edwards25519.NewScalar().SetCanonicalBytes(val)
		if err != nil {
			return
		}
		C = append(C, scalar)
	}
	for _, val := range ringSig.R {
		var scalar *edwards25519.Scalar
		scalar, err = edwards25519.NewScalar().SetCanonicalBytes(val)
		if err != nil {
			return
		}
		R = append(R, scalar)
	}
	return
}

func (ringSig RingSignature) ImageToPoint() (*edwards25519.Point, error) {
	return edwards25519.NewIdentityPoint().SetBytes(ringSig.Image)
}

func (ringSig RingSignature) CheckSignatureValidity() bool {
	// Parse from byte values
	c, r, err := ringSig.CRToScalars()
	if err != nil {
		fmt.Println(err)
		return false
	}
	I, err := ringSig.ImageToPoint()
	if err != nil {
		fmt.Println(err)
		return false
	}
	// Prepare L, R arrays and scalar for sum of c values
	L, R := make([]*edwards25519.Point, len(c)), make([]*edwards25519.Point, len(c))
	sumCi := edwards25519.NewScalar()
	for i := 0; i < len(c); i++ {
		L[i] = edwards25519.NewIdentityPoint().Add(
			edwards25519.NewIdentityPoint().ScalarBaseMult(r[i]),
			edwards25519.NewIdentityPoint().ScalarMult(c[i], ringSig.Utxos[i].Keypair.P),
		)
		R[i] = edwards25519.NewIdentityPoint().Add(
			edwards25519.NewIdentityPoint().ScalarMult(r[i], HashPoint(ringSig.Utxos[i].Keypair.P)),
			edwards25519.NewIdentityPoint().ScalarMult(c[i], I),
		)
		sumCi = edwards25519.NewScalar().Add(sumCi, c[i])
	}

	challenge, err := ComputeChallenge(ringSig.Utxos, L, R)
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Printf("h: %x  c: %x\n", challenge.Bytes(), sumCi.Bytes())
	if challenge.Equal(sumCi) != 1 {
		return false
	}

	return true
}
