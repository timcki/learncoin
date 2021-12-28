package main

import (
	"fmt"
	"math/rand"
	"time"

	"filippo.io/edwards25519"
	"github.com/timcki/learncoin/internal/chain"
	"github.com/timcki/learncoin/internal/transaction"
)

type ChainSimulation struct {
	Addr    []transaction.Address
	utxoSet chain.UtxoSet
}

func NewChainSimulation(addrQuant int, utxoSetSize int) *ChainSimulation {
	chainSim := ChainSimulation{utxoSet: chain.NewUtxoSet()}
	var addr []transaction.Address
	rand.Seed(time.Now().Unix())
	// Randomly generate addrQuant addresses
	for i := 0; i < addrQuant; i++ {
		a, _ := transaction.NewAddress()
		addr = append(addr, a)
	}
	// Randomly generate starting utxo set
	for i := 0; i < utxoSetSize; i++ {
		// Choose random address to generate one time key from
		randAddr := rand.Intn(addrQuant)
		dest, err := addr[randAddr].NewDestinationAddress()
		if err != nil {
			panic(err)
		}
		// Random value in (0, utxoSetSize/100)
		amt := float32(rand.Intn(utxoSetSize)) / 100
		utxo := transaction.NewUtxo(amt, dest)
		chainSim.utxoSet.Add(*utxo)
	}
	chainSim.Addr = addr
	return &chainSim
}

func (sim *ChainSimulation) scanAddress(n int) {
	for _, v := range sim.utxoSet.GetUtxos() {
		if sim.Addr[n].CheckDestinationAddress(v.Keypair) {
			fmt.Printf("Found linked utxo: %s\n", v.Bytes())
		}
	}
}

func main() {

	addr1, _ := transaction.NewAddress()
	addr2, _ := transaction.NewAddress()

	fmt.Println(addr1.PubKey.ToHumanReadable(false))
	fmt.Println(addr2.PubKey.ToHumanReadable(false))

	for i := 0; i < 3; i++ {
		if dest, err := addr1.NewDestinationAddress(); err != nil {
			panic(err)
		} else {
			check1 := addr1.CheckDestinationAddress(dest)
			check2 := addr2.CheckDestinationAddress(dest)
			if !check1 || check2 {
				panic("Failed check")
			}

			fmt.Printf("Checking for right address: %v\n", check1)
			fmt.Printf("Checking for wrong address: %v\n\n", check2)

			x, err := addr1.ComputePrivateKey(dest)
			if err != nil {
				panic(err)
			}
			P := new(edwards25519.Point).ScalarBaseMult(x)
			if P.Equal(dest.P) != 1 {
				panic("Couldn't recover tx priv key x")
			}
		}
	}

	var txns []transaction.Utxo

	kp, err := addr1.NewDestinationAddress()
	if err != nil {
		panic(err)
	}
	trueUtxo := transaction.NewUtxo(0, kp)
	for i := 0; i < 9; i++ {
		k, err := addr2.NewDestinationAddress()
		if err != nil {
			panic(err)
		}
		txns = append(txns, *transaction.NewUtxo(0, k))

	}
	ringSig := addr1.NewRingSignature(*trueUtxo, txns)
	fmt.Println(ringSig.CheckSignatureValidity())
	fmt.Println("Test passed")

	sim := NewChainSimulation(100, 30000)
	sim.scanAddress(10)
}
