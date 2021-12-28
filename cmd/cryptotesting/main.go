package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"

	"github.com/timcki/learncoin/internal/chain"
	"github.com/timcki/learncoin/internal/crypto"
	"github.com/timcki/learncoin/internal/transaction"
)

const RINGSIZE = 4

type ChainSimulation struct {
	Addr        []transaction.Address
	utxoSet     chain.UtxoSet
	utxoForAddr map[int][]crypto.FixedHash
}

func NewChainSimulation(addrQuant int, utxoSetSize int) *ChainSimulation {
	chainSim := ChainSimulation{
		utxoSet:     chain.NewUtxoSet(),
		utxoForAddr: make(map[int][]crypto.FixedHash),
	}
	var addr []transaction.Address
	rand.Seed(time.Now().Unix())
	// Randomly generate addrQuant addresses
	for i := 0; i < addrQuant; i++ {
		a, _ := transaction.NewAddress()
		addr = append(addr, a)
	}
	fmt.Printf("Randomized %d addresses...\n", addrQuant)
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
	fmt.Printf("Randomized %d utxos for those addresses...\n", utxoSetSize)
	chainSim.Addr = addr
	return &chainSim
}

func (sim *ChainSimulation) RandomTxn() {
	// Choose address at random
	a := rand.Intn(len(sim.Addr))
	addr := sim.Addr[a]
	if len(sim.utxoForAddr[a]) == 0 {
		num := sim.scanAddress(a)
		humanReadable, _ := addr.PubKey.ToHumanReadable(false)
		fmt.Printf("Found %d linked utxos for addr: %s\n", num, humanReadable)
	}

	// Pick random utxo and address to perform transaction
	// Utxo
	randUtxoNum := rand.Intn(len(sim.utxoForAddr[a]))
	trueUtxoHash := sim.utxoForAddr[a][randUtxoNum]
	trueUtxo := sim.utxoSet.Get(trueUtxoHash)

	decoyUtxos := make([]transaction.Utxo, 0)
	// Pick RINGSIZE-1 other txns with same amount
	trueHash, _ := trueUtxo.Hash()
	// There is a possibility that we'll find less than target ringsize
	// that's acceptable and we should continue as long as ringsize > 0
	for _, utxo := range sim.utxoSet.GetUtxos() {
		if len(decoyUtxos) == RINGSIZE-1 {
			break
		}
		decoyHash, _ := utxo.Hash()
		if utxo.Amount == trueUtxo.Amount && bytes.Compare(trueHash, decoyHash) != 0 {
			decoyUtxos = append(decoyUtxos, *utxo)
		}

	}
	if len(decoyUtxos) == 0 {
		fmt.Println("Not enough utxos with mirror amount, skipping")
		return
	}

	// Address
	var randAddrNum int
	for {
		randAddrNum = rand.Intn(len(sim.Addr))
		if randAddrNum != a {
			break
		}
	}
	addr2 := sim.Addr[randAddrNum]
	fmt.Println("Picked random destination address...")

	randomAmount := float32(rand.Intn(int(trueUtxo.Amount)*100)) / 100
	txn := addr.NewTransaction(append(decoyUtxos, *trueUtxo), randomAmount, addr2)
	message := txn.Bytes()
	fmt.Printf("Created new transaction: %s\n", message)

	fmt.Println("Computing ring signature for:")
	fmt.Printf("  Real utxo:    %s", trueUtxo.Bytes())
	for i, u := range decoyUtxos {
		fmt.Printf("  Decoy utxo %d: %s", i, u.Bytes())
	}

	ringSig := addr.NewRingSignature(*trueUtxo, decoyUtxos, txn.Bytes())
	fmt.Printf("Ring signature validation for true message:  %v\n", ringSig.CheckSignatureValidity(txn.Bytes()))
	fmt.Printf("Ring signature validation for false message: %v\n", ringSig.CheckSignatureValidity([]byte("Fake")))

	txn.Sigature = ringSig

	fmt.Printf("Signed transaction: %s\n", txn.Bytes())

}

// scanAddress scans the utxo set for utxos generated from own public keypair. Returns number of utxos founds
func (sim *ChainSimulation) scanAddress(n int) int {
	sum := 0
	for _, v := range sim.utxoSet.GetUtxos() {
		if sim.Addr[n].CheckDestinationAddress(v.Keypair) {
			sum += 1
			h, _ := v.Hash()
			sim.utxoForAddr[n] = append(sim.utxoForAddr[n], h.ToFixedHash())
		}
	}
	return sum
}

func main() {

	/*
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
	*/

	/*
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
	*/
	//ringSig := addr1.NewRingSignature(*trueUtxo, txns, message)
	//fmt.Println(ringSig.CheckSignatureValidity())
	//fmt.Println("Test passed")

	fmt.Println("====== learncoin chain simulation ======")
	fmt.Println(" This binary will generate a simulated")
	fmt.Println("     chain state and perform random")
	fmt.Println("     transactions every 2 seconds")
	fmt.Println("========================================")
	fmt.Print("\n\n====== Generating new chain simulation ======\n\n")
	sim := NewChainSimulation(1000, 150000)
	for {
		fmt.Println("\n==== Simulating transaction ====")
		sim.RandomTxn()
		time.Sleep(time.Second * 2)
	}
}
