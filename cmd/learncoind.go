package main

import (
	"encoding/gob"
	"encoding/json"
	"flag"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/timcki/learncoin/internal/config"
	"github.com/timcki/learncoin/internal/crypto"
	"github.com/timcki/learncoin/internal/messages"
	"github.com/timcki/learncoin/internal/node"
	"github.com/timcki/learncoin/internal/peer"
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

type Chain struct {
	blocks []Block
	mu     sync.RWMutex
}


/*
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
*/

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	gob.Register(messages.VersionMessage{})
	gob.Register(messages.VerAckMessage{})
	gob.Register(messages.PingMessage{})
	gob.Register(messages.PongMessage{})

}

func main() {

	// Parse CLI flags
	var defaultPeers bool
	var debugLogging bool
	var connPort string

	flag.StringVar(&connPort, "p", "", "Listener port")
	flag.BoolVar(&defaultPeers, "i", false, "Ignore default peer list on startup")
	flag.BoolVar(&debugLogging, "d", false, "Set log level to debug")
	flag.Parse()

	if debugLogging {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// testMerkleTree()

	// Read NodeConfig from disk or generate new one is non-existing
	var nodeConfig config.NodeConfig
	conf, err := os.Open("data/config.gob")
	if err != nil {
		log.Error().Err(err).Msg("Failed to read node config from disk")
		nodeConfig, err = config.NewNodeConfig()
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate node config exiting")
			os.Exit(-1)
		}
	} else {
		json.NewDecoder(conf).Decode(nodeConfig)
	}
	// Check port on which to launch connections
	if connPort != "" {
		if connPort == "random" {
			connPort = strconv.Itoa(8000 + rand.Intn(3000))
		}
		nodeConfig.SetPort(connPort)
	}

	node := node.NewNode(nodeConfig)

	// Default peer list
	// TODO: Move to file
	peerAddrs := []string{"localhost:8081"}

	// Connect to peers from the default list
	// TODO: Have only one node here and use the propagation algorithm
	// NOTE: In this way we make duplicate connections because
	// inbound port != outboud port
	if defaultPeers {
		for _, peerAddr := range peerAddrs {
			if err := node.NewOutboundPeer(peerAddr); err != nil {
				log.Error().Err(err).Str("peer", peerAddr).Msg("Failed connecton to peer")
			}
		}

	}

	node.Start()


}
