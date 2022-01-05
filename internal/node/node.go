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

type Node struct {
	config config.NodeConfig
	logger log.Logger
	peers  map[crypto.FixedHash]peer.Peer
}

func (n Node) GetPeers() map[crypto.FixedHash]peer.Peer {
	return n.peers
}

func (n Node) GetPeer(id crypto.FixedHash) peer.Peer {
	return n.peers[id]
}

func (n *Node) AddPeer(p peer.Peer) {
	n.peers[p.GetID()] = p
}

// Connects to a new peer (sends CmdVersion and waits for CmdVerAck)
func (n *Node) NewOutboundPeer(address string) (err error) {
	var conn net.Conn
	//n.logger.New()
	p := peer.NewPeer(n.logger.New("peer", address), n.getOtherPeers, n.NewOutboundPeer)
	conn, err = net.Dial(constants.ConnType, address)
	if err != nil {
		//n.logger.Error().Err(err).Str("addr", address).Msg("Failed connection to peer")
		return
	}
	p.SetConn(conn)

	msg := messages.NewVersionMessage(n.config.GetVersion(), n.config.GetAddr().ToString(), p.GetID())
	if err = p.WriteMessage(msg); err != nil {
		panic(err)
	}

	if err = p.HandleVersionMessage(); err != nil {
		return
	}

	p.SetInbound(false)
	p.SetAlive(true)

	// Add peer to peerlist and start inbound and outbound connections on it
	n.AddPeer(p)
	p.Start()
	n.logger.Info("Succesfully registered outbound peer", "peer", p.GetAddr().ToString())

	return err
}

// NewInboundPeer handles the connection of a new peer
func (n *Node) NewInboundPeer(conn net.Conn) (err error) {
	p := peer.NewPeer(n.logger.New("peer", conn.RemoteAddr().String()), n.getOtherPeers, n.NewOutboundPeer)
	p.SetConn(conn)

	var msg messages.Message
	if msg, err = p.ReadMessage(); err != nil {
		return
	}

	if msg.Command() == messages.CmdVersion {
		ver := msg.(messages.VersionMessage)
		port := config.NewAddressFromString(ver.Address).Port
		addr := config.NewAddressFromString(p.GetConn().RemoteAddr().String()).Addr
		finalAddr := config.NewAddress(addr, port)
		p.SetAddr(finalAddr)
	}

	msgg := messages.NewVersionMessage(n.config.GetVersion(), n.config.GetAddr().ToString(), p.GetID())
	if err = p.WriteMessage(msgg); err != nil {
		n.logger.Error("Failed send to peer", "err", err)
		return
	}
	p.SetInbound(true)
	p.SetAlive(true)

	p.Start()
	n.AddPeer(p)
	n.logger.Info("Got new inbound peer", "peer", p.GetAddr().ToString())
	return
}

// getOtherPeers creates a list of all known addresses (except for peer calling the callback)
// and returns it a slice
func (n Node) getOtherPeers(id crypto.FixedHash) []string {
	trueList := make([]string, 0)
	for _, p := range n.GetPeers() {
		if p.GetID() != id {
			trueList = append(trueList, p.GetAddr().ToString())
		}
	}
	return trueList
}

// Start starts listening for new connections on the port specified
// in the config file
func (n *Node) Start() {

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
			n.logger.Info("Got new inbound connection")
			if err := n.NewInboundPeer(conn); err != nil {
				n.logger.Error("Error while accepting connection", "err", err)
			} else {
			}
		}
	}

}

func NewNode(config config.NodeConfig, logger log.Logger) *Node {
	return &Node{config: config, logger: logger, peers: make(map[crypto.FixedHash]peer.Peer)}
}
