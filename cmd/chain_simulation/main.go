package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"

	"github.com/fatih/color"
	"github.com/timcki/learncoin/internal/chain"
	"github.com/timcki/learncoin/internal/crypto"
	"github.com/timcki/learncoin/internal/transaction"
)

const RINGSIZE = 8

var trueFalse = map[bool]string{
	true:  color.GreenString("✓"),
	false: color.RedString("𐄂"),
}

type ChainSimulation struct {
	// Simulation parameters
	Addr        []transaction.Address
	utxoSet     chain.UtxoSet
	utxoForAddr map[int][]crypto.FixedHash
	keyImages   [][]byte
	utxoValue   float32

	// Chain
	Chain chain.Chain
}

func NewChainSimulation(addrQuant int, utxoSetSize int) *ChainSimulation {
	chainSim := ChainSimulation{
		utxoSet:     chain.NewUtxoSet(),
		utxoForAddr: make(map[int][]crypto.FixedHash),
		Chain:       *chain.NewChain(),
		utxoValue:   float32(rand.Intn(utxoSetSize)/10) / 100,
		keyImages:   make([][]byte, 0),
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
		// Random value in (0, utxoSetSize/1000)
		amt := float32(rand.Intn(int(chainSim.utxoValue*100))) / 100
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

	// Pick a random amount for the transaction
	randomAmount := float32(rand.Intn(int(trueUtxo.Amount)*100)) / 100
	txn := addr.NewTransaction(append(decoyUtxos, *trueUtxo), randomAmount, addr2)
	fmt.Println("Created new transaction...")

	fmt.Println("Computing ring signature for transaction with:")
	fmt.Printf("  Real utxo:    %s", trueUtxo.Bytes())
	for i, u := range decoyUtxos {
		fmt.Printf("  Decoy utxo %d: %s", i, u.Bytes())
	}

	// Message is the byte representation of our txn
	message := txn.Bytes()

	// Sign the message with trueUtxo+decoyUtxos
	ringSig := addr.NewRingSignature(*trueUtxo, decoyUtxos, message)
	fmt.Println("\nSigned byte representation of transaction")
	fmt.Printf("%s:\n", color.BlueString("Ring signature validation"))
	fmt.Printf("  Transaction:  %v\n", trueFalse[ringSig.CheckSignatureValidity(message)])
	fmt.Printf("  Fake message: %v\n", trueFalse[ringSig.CheckSignatureValidity([]byte("Fake"))])

	// Assign the ring signature to our txns
	txn.Sigature = ringSig
	// Append of txn to the used key images set to prevent double spending
	sim.keyImages = append(sim.keyImages, ringSig.Image)
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

	fmt.Printf("\n====== %s ======\n", color.GreenString("learncoin chain simulation"))
	fmt.Println(" This binary will generate a simulated")
	fmt.Println("     chain state and perform random")
	fmt.Println("      transactions every 2 seconds")
	fmt.Println("========================================")
	fmt.Printf("\n\n====== %s ======\n\n", color.BlueString("Generating new chain simulation"))
	sim := NewChainSimulation(1000, 150000)

	mempool := make([]crypto.Hashable, 0)
	for {
		fmt.Printf("\n==== %s ====\n", color.BlueString("Simulating transaction"))
		txn := sim.RandomTxn()
		if txn != nil {
			fmt.Printf("Signed transaction: %s\n", txn.PrettyPrint())
			mempool = append(mempool, txn)
		}
		// Construct block with 50% prob if more than two txns
		if len(mempool) > 2 && rand.Intn(2) < 1 {
			fmt.Printf("\n\n====== %s ======\n\n", color.BlueString("Constructing block from transactions"))
			block := chain.NewBlock(mempool)
			sim.Chain.AddBlock(block)
			fmt.Printf("%s\n", color.BlueString("Added block to chain"))
			fmt.Printf("%s: %s\n", color.YellowString("header"), block.Header.PrettyPrint())
			fmt.Printf("%s:\n%s\n", color.YellowString("txns"), block.Transactions.PrettyPrint())
			fmt.Printf("\n%s: %d\n", color.BlueString("Chain length"), sim.Chain.Length())
			fmt.Printf("%s\n", color.BlueString("Clearing mempool"))
			mempool = make([]crypto.Hashable, 0)
		}
		time.Sleep(time.Second * 2)
	}
}
