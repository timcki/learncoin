package config

import (
	"encoding/binary"
	"math/rand"
	"time"

	"github.com/timcki/learncoin/internal/constants"
	"github.com/timcki/learncoin/internal/crypto"
)

// NodeConfig holds all of the important configuration related to the node's inner workings
// i.e. private/public keys, identity hash
/*type NodeConfig interface {
	GetID() crypto.Hash
	GetPeers() map[crypto.FixedHash]peer.Peer
	GetPeer(crypto.FixedHash) peer.Peer

	AddPeer(crypto.FixedHash, peer.Peer)
}
*/

type NodeConfig struct {
	// Connection vars
	port     string
	connType string
	connAddr string
	// Protocol vars
	version string
	id      crypto.Hash
}

func (c *NodeConfig) SetPort(p string) {
	c.port = p
}

func (c *NodeConfig) GetPort() string {
	return c.port
}

func (c *NodeConfig) GetConnType() string {
	return c.connType
}

func (c *NodeConfig) GetConnAddr() string {
	return c.connAddr
}

func (c *NodeConfig) GetID() crypto.Hash {
	return c.id
}

func (c *NodeConfig) GetVersion() string {
	return c.version
}

func generateNewIdentity() (crypto.Hash, error) {
	var d []byte
	var nonce []byte
	if d, err := time.Now().MarshalText(); err != nil {
		return d, err
	}
	binary.BigEndian.PutUint64(nonce, uint64(rand.Int63()))
	return crypto.HashData(append(nonce, d...))
}

func NewNodeConfig() (NodeConfig, error) {
	var err error
	// Set default port as 8080
	c := NodeConfig{
		port:     "8080",
		connType: constants.ConnType,
		connAddr: constants.ConnAddr,
		version:  constants.Version,
	}

	c.id, err = generateNewIdentity()
	if err != nil {
		return c, err
	}
	return c, err
}
