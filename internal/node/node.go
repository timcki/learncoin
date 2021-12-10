package node

import (
	"net"

	log "github.com/inconshreveable/log15"
	"github.com/timcki/learncoin/internal/config"
	"github.com/timcki/learncoin/internal/constants"
	"github.com/timcki/learncoin/internal/crypto"
	"github.com/timcki/learncoin/internal/messages"
	"github.com/timcki/learncoin/internal/peer"
)

type Node interface {
	// Peer related read/write functions
	GetPeers() map[crypto.FixedHash]peer.Peer
	GetPeer(crypto.FixedHash) peer.Peer
	AddPeer(peer.Peer)

	NewOutboundPeer(string) error
	NewInboundPeer(net.Conn) error
	// Start accepting incomming connections

	Start()
}

type node struct {
	config config.NodeConfig

	logger log.Logger

	peers map[crypto.FixedHash]peer.Peer
}

func (n node) GetPeers() map[crypto.FixedHash]peer.Peer {
	return n.peers
}

func (n node) GetPeer(id crypto.FixedHash) peer.Peer {
	return n.peers[id]
}

func (n node) AddPeer(p peer.Peer) {
	n.peers[p.GetID()] = p
	n.logger.Debug("Succesfully registered peer", "peer", p.GetAddr().ToString())
}

// Connects to a new peer (sends CmdVersion and waits for CmdVerAck)
func (n node) NewOutboundPeer(address string) (err error) {
	var conn net.Conn
	p := peer.NewPeer(n.logger.New("peer", address))
	conn, err = net.Dial(constants.ConnType, address)
	if err != nil {
		//n.logger.Error().Err(err).Str("addr", address).Msg("Failed connection to peer")
		return
	}
	p.SetConn(conn)

	msg := messages.NewVersionMessage(n.config.GetVersion(), n.config.GetAddr().ToString(), p.GetID())
	if err = p.SendVersionMessage(msg); err != nil {
		return
	}
	if err = p.HandleVersionMessage(); err != nil {
		return
	}

	p.SetInbound(false)
	p.SetAlive(true)

	// Add peer to peerlist and start inbound and outbound connections on it
	n.AddPeer(p)
	p.Start()
	n.logger.Debug("Succesfully registered outbound peer", "peer", p.GetAddr().ToString())

	return err
}

// NewInboundPeer handles the connection of a new peer
func (n node) NewInboundPeer(conn net.Conn) (err error) {
	p := peer.NewPeer(n.logger.New("peer", conn.RemoteAddr().String()))
	p.SetConn(conn)
	if err = p.HandleVersionMessage(); err != nil {
		n.logger.Error("Failed connection to peer", "err", err)
		return
	}
	msg := messages.NewVersionMessage(n.config.GetVersion(), n.config.GetAddr().ToString(), p.GetID())
	if err = p.SendVersionMessage(msg); err != nil {
		n.logger.Error("Failed send to peer", "err", err)
		return
	}
	p.SetInbound(true)
	p.SetAlive(true)

	n.AddPeer(p)
	n.logger.Info("Got new inbound peer", "peer", p.GetAddr().ToString())
	return
}

func (n node) Start() {

	listener, err := net.Listen(n.config.GetConnType(), n.config.GetAddr().ToString())
	if err != nil {
		n.logger.Error("Error while opening listener", "err", err)
		panic(err)
	}
	defer listener.Close()
	n.logger.Info("Started server", "addr", n.config.GetAddr().ToString())

	for {
		if conn, err := listener.Accept(); err != nil {
			n.logger.Error("Error while accepting connection", "err", err)
		} else {
			if err := n.NewInboundPeer(conn); err != nil {
				n.logger.Error("Error while accepting connection", "err", err)
			} else {
			}
		}
	}

}

func NewNode(config config.NodeConfig, logger log.Logger) Node {
	return node{config: config, logger: logger}
}
