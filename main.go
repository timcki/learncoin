package main

import (
	"encoding/gob"
	"flag"
	"net"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/timcki/learncoin/crypto"
	"github.com/timcki/learncoin/messages"
)

// Header is the header of a block
type Header struct {
	version      uint8
	previousHash crypto.Hash
	merkleRoot   crypto.Hash
	time         time.Time
}

type Block struct {
	header       Header
	transactions crypto.MerkleTree
}

type BlockChain struct {
	blocks []Block
	mu     sync.RWMutex
}

// TODO: Move this to the peer package?
const (
	connHost string = "localhost"
	connType string = "tcp"
)

// TODO: Persistent storage of known nodes
var (
	connPort string
	Peers    = make(map[string]*Peer)
)

func testMerkleTree() {
	txs := []crypto.Hashable{randomTransaction(), randomTransaction(), randomTransaction()}
	for _, tx := range txs {
		hash, _ := tx.Hash()
		log.Info().Msg(hash.String())
	}
	tree, err := crypto.NewMerkleTree(txs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initiate merkle tree: ")
		return
	}
	for i, el := range tree.GetNodes() {
		log.Info().Msgf("%d %+v", i, el)
	}
	log.Info().Msg(tree.RootHash().String())
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// testMerkleTree()

	// Parse CLI flags
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
	peerAddrs := []string{"localhost:8081"}

	gob.Register(messages.VersionMessage{})
	gob.Register(messages.VerAckMessage{})
	gob.Register(messages.PingMessage{})
	gob.Register(messages.PongMessage{})

	// Connect to peers from the default list
	// TODO: Have only one node here and use the propagation algorithm
	// NOTE: In this way we make duplicate connections because
	// inbound port != outboud port
	if defaultPeers {
		for _, peerAddr := range peerAddrs {
			if peer, err := NewOutboundPeer(peerAddr); err != nil {
				log.Error().Err(err).Str("peer", peerAddr).Msg("Failed connection to peer")
			} else {
				peer.start()
				Peers[peer.addr] = peer
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
			if peer, err := NewInboundPeer(conn); err != nil {
				log.Error().Err(err).Msg("Error while accepting connection")
			} else {
				log.Info().Msgf("Got peer from %s", peer.addr)
				peer.start()
				Peers[peer.addr] = peer
			}
		}
	}

}
