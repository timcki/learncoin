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
	// Simulation parameters
	Addr        []transaction.Address
	utxoSet     chain.UtxoSet
	utxoForAddr map[int][]crypto.FixedHash

	// Chain
	Chain chain.Chain
}

func NewChainSimulation(addrQuant int, utxoSetSize int) *ChainSimulation {
	chainSim := ChainSimulation{
		utxoSet:     chain.NewUtxoSet(),
		utxoForAddr: make(map[int][]crypto.FixedHash),
		Chain:       *chain.NewChain(),
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

func (sim *ChainSimulation) RandomTxn() *transaction.Transaction {
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
		return nil
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
	fmt.Println("Created new transaction...")

	fmt.Println("Computing ring signature for:")
	fmt.Printf("  Real utxo:    %s", trueUtxo.Bytes())
	for i, u := range decoyUtxos {
		fmt.Printf("  Decoy utxo %d: %s", i, u.Bytes())
	}

	ringSig := addr.NewRingSignature(*trueUtxo, decoyUtxos, txn.Bytes())
	fmt.Printf("Ring signature validation for true message:  %v\n", ringSig.CheckSignatureValidity(txn.Bytes()))
	fmt.Printf("Ring signature validation for false message: %v\n", ringSig.CheckSignatureValidity([]byte("Fake")))

	txn.Sigature = ringSig

	return &txn
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

	fmt.Println("====== learncoin chain simulation ======")
	fmt.Println(" This binary will generate a simulated")
	fmt.Println("     chain state and perform random")
	fmt.Println("      transactions every 2 seconds")
	fmt.Println("========================================")
	fmt.Print("\n\n====== Generating new chain simulation ======\n\n")
	sim := NewChainSimulation(1000, 150000)

	mempool := make([]crypto.Hashable, 0)
	for {
		fmt.Println("\n==== Simulating transaction ====")
		txn := sim.RandomTxn()
		if txn != nil {
			fmt.Printf("Signed transaction: %s\n", txn.PrettyPrint())
			mempool = append(mempool, txn)
		}
		// Construct block with 50% prob if more than two txns
		if len(mempool) > 2 && rand.Intn(2) < 1 {
			fmt.Println("Constructing block from transactions")
			block := chain.NewBlock(mempool)
			sim.Chain.AddBlock(*block)
			fmt.Printf("Added block to chain: %s\n", block.PrettyPrint())
			fmt.Printf("Chain length: %d\n", sim.Chain.Length())
			fmt.Println("Clearing mempool")
			mempool = make([]crypto.Hashable, 0)
		}
		time.Sleep(time.Second * 2)
	}
}
