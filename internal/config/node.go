package config

import (
	"encoding/binary"
	"math/rand"
	"strings"
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

type Address struct {
	Addr string
	Port string
}

func NewAddressFromString(addr string) Address {
	a := strings.Split(addr, ":")
	return Address{Addr: a[0], Port: a[1]}
}

func NewAddress(addr, port string) Address {
	return Address{Addr: addr, Port: port}
}

func (a Address) ToString() string {
	return a.Addr + ":" + a.Port
}

type NodeConfig struct {
	// Connection vars
	addr     Address
	connType string
	// Protocol vars
	version string
	id      crypto.Hash
}

func (c *NodeConfig) SetAddr(a Address) {
	c.addr = a
}

func (c *NodeConfig) GetAddr() Address {
	return c.addr
}

func (c *NodeConfig) GetConnType() string {
	return c.connType
}

func (c *NodeConfig) GetID() crypto.Hash {
	return c.id
}

func (c *NodeConfig) GetVersion() string {
	return c.version
}

func generateNewIdentity() (crypto.Hash, error) {
	var d []byte
	nonce := make([]byte, 8)
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
		addr:     NewAddress(constants.ConnAddr, "8080"),
		connType: constants.ConnType,
		version:  constants.Version,
	}

	c.id, err = generateNewIdentity()
	if err != nil {
		return c, err
	}
	return c, err
}
