package main

import (
	"bufio"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"log"
	"math"
	"math/rand"
	"net"
	"os"
	"time"
)

// Hashable guarantees that a given type implements
// the Hash function
type Hashable interface {
	Hash() (Hash, error)
}

// Hash is a convenience type used for representing a
// hash digest
type Hash []byte

// Header is the header of a block
type Header struct {
	version       uint8
	previous_hash Hash
	merkle_root   Hash
	time          time.Time
}

func hash(data []byte) (Hash, error) {
	sum := sha512.New()
	if _, err := sum.Write(data); err != nil {
		return nil, err
	}
	return sum.Sum(nil), nil
}

func (h Hash) String() string {
	return hex.EncodeToString(h)
}

type Node struct {
	leaf    bool
	Hash    Hash
	Content *Hashable
}

type EmptyLeaf int

func (e EmptyLeaf) Hash() (Hash, error) {
	h := make([]byte, 0)
	return h, nil
}

// TODO: To speedup lookup time create a hashtable with mapping
// tx_hash -> array_index (maybe overoptimization?)
type MerkleTree struct {
	nodes []Node
	depth uint
}

// NewMerkleTree constructs a binary tree in which
// the leaves are transactions and parents are subsequent
// hashes. Since we know the depth at which the children will start
func NewMerkleTree(elements []Hashable) (*MerkleTree, error) {
	var pow float64
	elements, pow = fillElements(elements)
	// Tree size is sum(2^n for n from pow to 1)
	// so also 2^(n+1) - 1
	size := int(math.Pow(2, pow+1) - 1)
	tree := make([]Node, size)

	// Calculate the base index for level of leaves
	baseIndex := int(math.Pow(2, pow)) - 1
	// Fill all of the leaves with txs
	for i, element := range elements {
		hash, err := element.Hash()
		if err != nil {
			return nil, err
		}
		tree[baseIndex+i] = Node{leaf: true, Hash: hash, Content: &element}
	}
	// Fill all parent nodes until root
	for i := baseIndex - 1; i >= 0; i-- {
		childrenHash, err := hash(append(tree[2*i+1].Hash, tree[2*i+2].Hash...))
		if err != nil {
			return nil, err
		}
		tree[i] = Node{leaf: false, Hash: childrenHash, Content: nil}
	}
	return &MerkleTree{tree, uint(baseIndex)}, nil
}

// fillElements rounds the number of leafs to the nearest
// power of 2 greater than the length
func fillElements(el []Hashable) ([]Hashable, float64) {
	l := len(el)
	// Calculate the nearest power of two greater than the len of our tx list
	pow := math.Ceil(math.Log2(float64(l)))
	for i := 0; i < int(math.Pow(2, pow))-l; i++ {
		el = append(el, EmptyLeaf(0))
	}
	return el, pow
}

func (m *MerkleTree) RootHash() Hash {
	return m.nodes[0].Hash
}

// TODO: Print the MerkleTree in a pretty form with hashes and values
func (m *MerkleTree) PresentTree() string {
	//result := ""
	return ""
}

type Block struct {
	header       Header
	transactions MerkleTree
}

type Blockchain struct {
	blocks []Block
}

func HashBlock(block Block) Hash {
	return make([]byte, 4)
}

const (
	connHost = "localhost"
	connType = "tcp"
)

func main() {
	/*
		txs := []Hashable{&Transaction{12}, &Transaction{24}, &Transaction{36}}
		for _, tx := range txs {
			hash, _ := tx.Hash()
			fmt.Println(hash.String())
		}
		tree, err := NewMerkleTree(txs)
		if err != nil {
			panic(err)
		}
		for i, el := range tree.nodes {
			fmt.Println(i, el)
		}
		fmt.Println(tree.RootHash().String())
	*/

	// Define custom port as cli param
	var connPort string
	var defaultPeers bool
	flag.StringVar(&connPort, "p", "8080", "Listener port")
	flag.BoolVar(&defaultPeers, "i", false, "Ignore default peer list on startup")
	flag.Parse()
	// Default peer list
	peers := []string{"localhost:8081", "localhost:8082"}

	if defaultPeers {
		for _, peer := range peers {
			if conn, err := net.Dial(connType, peer); err != nil {
				log.Println("Failed connection to peer", peer, err.Error())
			} else {
				go connectToPeer(conn)
			}
		}

	}

	log.Printf("Starting %s server on %s:%s\n", connType, connHost, connPort)
	listener, err := net.Listen(connType, connHost+":"+connPort)
	if err != nil {
		log.Fatalln("Error while opening listener: ", err.Error())
		os.Exit(1)
	}

	defer listener.Close()

	for {
		if conn, err := listener.Accept(); err != nil {
			log.Fatalln("Error while accepting connection: ", err.Error())
		} else {
			log.Printf("Got peer from %s", conn.RemoteAddr().String())
			go handlePeerConnection(conn)
		}
	}

}

func connectToPeer(conn net.Conn) {
	for {
		tx := Transaction{rand.Intn(100)}
		conn.Write(tx.toBytes())
		conn.Write([]byte("\n"))
		time.Sleep(3 * time.Second)
	}
}

func handlePeerConnection(conn net.Conn) {
	for {
		buffer, err := bufio.NewReader(conn).ReadBytes('\n')
		if err != nil {
			log.Println("Terminating connection with client: ", err.Error())
			conn.Close()
			return
		}
		tx := binary.LittleEndian.Uint32(buffer[:len(buffer)-1])
		log.Printf("Got message: %v from peer %s", Transaction{int(tx)}, conn.RemoteAddr().String())
		//conn.Write([]byte("ACC\n"))
		//time.Sleep(3 * time.Second)
	}
}
