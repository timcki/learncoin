package main

import (
	"flag"
	"net"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Header is the header of a block
type Header struct {
	version      uint8
	previousHash Hash
	merkleRoot   Hash
	time         time.Time
}

type Block struct {
	header       Header
	transactions MerkleTree
}

type BlockChain struct {
	blocks []Block
	mu     sync.RWMutex
}

const (
	connHost = "localhost"
	connType = "tcp"
)

var (
	Peers = make(map[string]*Peer)
)

func testMerkleTree() {
	txs := []Hashable{randomTransaction(), randomTransaction(), randomTransaction()}
	for _, tx := range txs {
		hash, _ := tx.Hash()
		log.Info().Msg(hash.String())
	}
	tree, err := NewMerkleTree(txs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initiate merkle tree: ")
		return
	}
	for i, el := range tree.nodes {
		log.Info().Msgf("%d %+v", i, el)
	}
	log.Info().Msg(tree.RootHash().String())
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// testMerkleTree()

	// Parse CLI flags
	var connPort string
	var defaultPeers bool
	var debugLogging bool
	flag.StringVar(&connPort, "p", "8080", "Listener port")
	flag.BoolVar(&defaultPeers, "i", false, "Ignore default peer list on startup")
	flag.BoolVar(&debugLogging, "d", false, "Set log level to debug")
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debugLogging {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Default peer list
	peerAddrs := []string{"localhost:8081", "localhost:8082"}

	// Connect to peers from the default list
	// TODO: Have only one node here and use the propagation algorithm
	if defaultPeers {
		for _, peerAddr := range peerAddrs {
			if peer, err := newOutboundPeer(peerAddr); err != nil {
				log.Error().Err(err).Str("peer", peerAddr).Msg("Failed connection to peer")
			} else {
				peer.start()
				Peers[peerAddr] = peer
			}
		}

	}

	log.Info().Msgf("Starting %s server on %s:%s", connType, connHost, connPort)

	listener, err := net.Listen(connType, connHost+":"+connPort)
	if err != nil {
		log.Fatal().Err(err).Msg("Error while opening listener")
	}
	defer listener.Close()

	for {
		if conn, err := listener.Accept(); err != nil {
			log.Error().Err(err).Msg("Error while accepting connection")
		} else {
			addr := conn.RemoteAddr().String()
			if peer, err := newInboundPeer(conn); err != nil {
				log.Error().Err(err).Msgf("Error while accepting connection from %s", addr)
			} else {
				log.Info().Msgf("Got peer from %s", addr)
				peer.start()
				Peers[addr] = peer
			}
		}
	}

}
