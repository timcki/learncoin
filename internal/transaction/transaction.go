package transaction

import (
	"bytes"
	"encoding/json"
	"fmt"

	//"github.com/TylerBrock/colorjson"
	"github.com/timcki/learncoin/internal/crypto"
)

// Address is a mock structure for representing addresses
// TODO: Implement valid addresses

// Utxo represents an unspent transaction output that's used in actual transactions to move value between addresses

type Utxo struct {
	hash    crypto.Hash
	Amount  float32
	Keypair OneTimeAddress
}

// Transaction represents a transaction in the learncoin network. It has one ring input and returns either a single output
// or the output and change back to the origin address
type Transaction struct {
	UtxosIn  []Utxo
	UtxosOut []Utxo
	To       OneTimeAddress
	Sigature RingSignature
}

// CheckValidity performs checks making sure that the txn is valid
func (t Transaction) CheckValidity() bool {
	var sumIn float32 = t.UtxosIn[0].Amount
	var sumOut float32 = 0
	// check if all amounts are >0
	for _, utxo := range t.UtxosIn {
		// Check if all inputs in ring sig have the same amount
		if utxo.Amount < 0 || utxo.Amount != sumIn {
			return false
		}
	}
	for _, utxo := range t.UtxosOut {
		if utxo.Amount < 0 {
			return false
		}
		sumOut += utxo.Amount
	}
	return sumIn == sumOut
}

func (utxo Utxo) Bytes() []byte {
	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(utxo)
	return buffer.Bytes()
}

func (utxo Utxo) CheckValidity() bool {
	hash, err := utxo.Hash()
	if err != nil {
		return false
	}
	return bytes.Compare(utxo.hash, hash) == 0
}

func (utxo Utxo) Hash() (crypto.Hash, error) {
	if len(utxo.hash) != 0 {
		return utxo.hash, nil
	}
	var buf bytes.Buffer
	amt := []byte(fmt.Sprintf("%x", utxo.Amount))
	buf.Write(amt)
	buf.Write(utxo.Keypair.P.Bytes())
	buf.Write(utxo.Keypair.R.Bytes())

	return crypto.HashData(buf.Bytes())
}

// NewUtxo returns
func NewUtxo(amount float32, keypair OneTimeAddress) *Utxo {
	utxo := Utxo{
		Amount:  amount,
		Keypair: keypair,
	}
	hash, err := utxo.Hash()
	if err != nil {
		panic(err)
	}
	utxo.hash = hash
	return &utxo
}

// Used for testing purposes

// I'm choosing the default json encoder for encoding. It increases interoperability
// while lossing out on space. This is a reference design.
// In case of a problem it's also abstracted away so easy to change
func (t Transaction) Bytes() []byte {
	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(t)
	return buffer.Bytes()
}

func (t Transaction) PrettyPrint() string {
	res, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
	return string(res)
}

func (t Transaction) Hash() (crypto.Hash, error) {
	if h, err := crypto.HashData(t.Bytes()); err != nil {
		return nil, err
	} else {
		return h, nil
	}
}
