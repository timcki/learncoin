package node

import (
	"net"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/timcki/learncoin/internal/config"
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

	logger zerolog.Logger

	peers map[crypto.FixedHash]peer.Peer
}

func (n node) GetPeers() map[crypto.FixedHash]peer.Peer {
	return n.peers
}

func (n node) GetPeer(id crypto.FixedHash) peer.Peer {
	return n.peers[id]
}

func (n node) AddPeer(p peer.Peer) {
	n.peers[p.GetID().ToFixedHash()] = p
	log.Debug().Str("peer", p.GetAddr()).Msg("Succesfully registered peer")
}

// Connects to a new peer (sends CmdVersion and waits for CmdVerAck)
func (n node) NewOutboundPeer(address string) (err error) {
	var conn net.Conn
	p := peer.NewPeer(n.logger.With().Logger())
	conn, err = net.Dial("tcp", address)
	if err != nil {
		n.logger.Error().Err(err).Str("addr", address).Msg("Failed connection to peer")
		return
	}
	p.SetConn(conn)

	msg := messages.NewVersionMessage(n.config.GetVersion(), p.GetID())
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
	log.Debug().Str("peer", p.GetAddr()).Msg("Succesfully registered outbound peer")

	return err
}

// NewInboundPeer handles the connection of a new peer
func (n node) NewInboundPeer(conn net.Conn) (err error) {
	p := peer.NewPeer(n.logger.With().Logger())
	p.SetConn(conn)
	if err = p.HandleVersionMessage(); err != nil {
		n.logger.Error().Err(err).Msg("Failed connection to peer")
		return
	}
	msg := messages.NewVersionMessage(n.config.GetVersion(), p.GetID())
	if err = p.SendVersionMessage(msg); err != nil {
		log.Error().Err(err).Msg("Failed send to peer")
		return
	}
	p.SetInbound(true)
	p.SetAlive(true)

	n.AddPeer(p)
	return
}

func (n node) Start() {
	//log.Info().Msgf("Starting %s server on %s:%s", connType, connHost, connPort)

	listener, err := net.Listen(n.config.GetConnType(), n.config.GetConnAddr()+":"+n.config.GetPort())
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

func NewNode(config config.NodeConfig) Node {
	return node{config: config}
}
