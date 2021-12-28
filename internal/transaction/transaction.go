package transaction

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

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

type Transaction struct {
	UtxosIn  [][]*Utxo
	UtxosOut []*Utxo
	From     Address
	To       Address
}

// CheckValidity performs checks making sure that the txn is valid
func (t Transaction) CheckValidity() bool {
	var sumIn float32 = 0
	var sumOut float32 = 0
	// check if all amounts are >0
	for _, utxos := range t.UtxosIn {
		// Check if all inputs in ring sig have the same amount
		val := utxos[0].Amount
		if val < 0 {
			return false
		}
		for _, utxo := range utxos[1:] {
			if val != utxo.Amount || utxo.Amount < 0 {
				return false
			}
		}
		sumIn += val
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
	//fmt.Printf(string(buffer.Bytes()))
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
//func randomTransaction() *Transaction {
//sender := Address{rand.Intn(16384), rand.Intn(16384)}
//receiver := Address{rand.Intn(16384), rand.Intn(16384)}
//return &Transaction{sender, receiver, rand.Intn(1000)}
//}

//func NewTransaction(sender, receiver Address, amount int) Transaction {
//return Transaction{Sender: sender, Receiver: receiver, Amount: amount}
//}

// I'm choosing the default Go byte encoder (gob). It reduces interoperability
// but it's not important since this is a reference design.
// In case of a problem it's also abstracted away so easy to change
func (t Transaction) Bytes() []byte {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	encoder.Encode(t)
	return buffer.Bytes()
}

func (t Transaction) Hash() (crypto.Hash, error) {
	if h, err := crypto.HashData(t.Bytes()); err != nil {
		return nil, err
	} else {
		return h, nil
	}
}
