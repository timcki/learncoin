package main

import (
	"encoding/gob"
	"encoding/json"
	"math/rand"
	"os"
	"strconv"

	log "github.com/inconshreveable/log15"
	"github.com/timcki/learncoin/internal/chain"
	"github.com/timcki/learncoin/internal/config"
	"github.com/timcki/learncoin/internal/constants"
	"github.com/timcki/learncoin/internal/messages"
	"github.com/timcki/learncoin/internal/node"
)

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

func initGob() {
	gob.Register(messages.VersionMessage{})
	gob.Register(messages.VerAckMessage{})
	gob.Register(messages.PingMessage{})
	gob.Register(messages.PongMessage{})
}

func testCrypto() {
	logger := log.New()
	address, err := chain.NewAddress()
	if err != nil {
		logger.Error("", "err", err)
	}
	logger.Info(address.PubKey.ToHumanReadable())
}

func main() {
	initGob()
	testCrypto()

	var logger log.Logger
	if os.Getenv("ENVIRONMENT") == "dev" {
		logger = log.New()
	} else {
		logger = log.New()
	}
	logger.SetHandler(log.MultiHandler(
		log.StreamHandler(os.Stderr, log.LogfmtFormat()),
	))

	// testMerkleTree()

	// Read NodeConfig from disk or generate new one is non-existing
	var nodeConfig config.NodeConfig
	conf, err := os.Open("data/config.json")
	if err != nil {
		logger.Warn("Failed to read node config from disk")
		nodeConfig, err = config.NewNodeConfig()
		if err != nil {
			logger.Error("Failed to generate node config", "err", err)
			os.Exit(-1)
		}
		logger.Info("Generated new node config")
	} else {
		json.NewDecoder(conf).Decode(&nodeConfig)
	}
	// Check port on which to launch connections
	connPort := os.Getenv("NODE_PORT")
	if connPort != "" {
		if connPort == "random" {
			connPort = strconv.Itoa(8000 + rand.Intn(3000))
		}
		nodeConfig.SetAddr(config.NewAddress(constants.ConnAddr, connPort))
	}

	node := node.NewNode(nodeConfig, logger.New("node", "main_node"))

	// Default peer list
	// TODO: Move to file
	//peerAddrs := []string{"bootstrapper:8080"}

	// Connect to peers from the default list
	// TODO: Have only one node here and use the propagation algorithm
	if os.Getenv("BOOTSTRAP_NODE") != "" {
		logger.Info("Starting bootstrap")
		if err := node.NewOutboundPeer(os.Getenv("BOOTSTRAP_NODE")); err != nil {
			logger.Error("Failed connecton to peer", "peer", os.Getenv("BOOTSTRAP_NODE"))
		}
	}

	node.Start()
}
