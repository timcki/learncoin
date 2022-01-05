package peer

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"math/rand"
	"net"
	"time"

	log "github.com/inconshreveable/log15"
	"github.com/timcki/learncoin/internal/config"
	"github.com/timcki/learncoin/internal/crypto"
	"github.com/timcki/learncoin/internal/messages"
)

const (
	connHost string = "localhost"
	connType string = "tcp"
)

var (
	NoVersionMessageOnInitError  = errors.New("Didn't receive VersionMessage on initial connection")
	MalformedVersionMessageError = errors.New("Malformed VersionMessage on initial connection")
)

type Peer struct {
	// Connection to the peer. From doc:
	// Multiple goroutines may invoke methods on a Conn simultaneously.
	// So no mutex needed
	conn net.Conn

	// Logger dedicated to the peer
	logger log.Logger

	// Safe for concurrent access, set at creation and won't change
	id      crypto.FixedHash
	addr    config.Address
	inbound bool

	// Stats that arrive from node with the version flag

	// Channels for internal message communication
	// network handler -> internal executor
	// NOTE: Abandoned idea for now, might be useful
	// while using more abstractions later on e.g. when
	// the outHandler and messageQueue will be separate goroutines
	//inChan  chan Message
	//outChan chan Message

	// Should be called from only one place so
	// safe for concurrent access (I think?)
	alive bool

	// Callbacks to node
	getPeers                func(crypto.FixedHash) []string
	newOutboundPeerCallback func(string) error
}

func (p *Peer) SetConn(conn net.Conn) {
	p.conn = conn
}

func (p *Peer) SetInbound(b bool) {
	p.inbound = b
}

func (p *Peer) SetAlive(b bool) {
	p.alive = b
}

func (p *Peer) SetAddr(addr config.Address) {
	p.addr = addr
}

func (p Peer) GetConn() net.Conn {
	return p.conn
}

func (p Peer) GetAddr() config.Address {
	return p.addr
}

func (p Peer) GetID() crypto.FixedHash {
	return p.id
}

func (p Peer) IsInboud() bool {
	return p.inbound
}
func (p Peer) IsAlive() bool {
	return p.alive
}

func (p Peer) ReadMessage() (messages.Message, error) {
	var msg messages.Message
	if err := gob.NewDecoder(p.conn).Decode(&msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func parseMessage[T messages.Msg](conn net.Conn) (*T, error) {
	msg := new(T)
	if err := json.NewDecoder(conn).Decode(msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (p Peer) WriteMessage(msg messages.Message) error {
	return gob.NewEncoder(p.conn).Encode(&msg)
}

func (p *Peer) HandleAddressMessage(msg messages.AddrMessage) {
	for _, addr := range msg.Nodes {
		if err := p.newOutboundPeerCallback(addr); err != nil {
			p.logger.Warn("Failed to connect to node", "addr", addr)
		} else {
			p.logger.Debug("Connected to", "addr", addr)
		}
	}
}

func (p *Peer) HandleGetAddressMessage() error {
	otherPeers := p.getPeers(p.id)
	msg := messages.NewAddrMessage(otherPeers)
	return p.WriteMessage(msg)
}

// handleVersionMessage reads a message from the peer network connection,
// parses it if it's a VersionMessage and sends a VerAck, otherwise returns error
func (p *Peer) HandleVersionMessage() error {
	msg, err := p.ReadMessage()
	if err != nil {
		return err
	}
	ver := msg.(messages.VersionMessage)

	address := p.conn.RemoteAddr().String()
	p.logger.Info("Remote addr string", "addr", address)
	p.addr = config.NewAddressFromString(address)
	p.id = ver.ID
	p.logger.Debug("addr", "addr", p.addr)

	return nil
}

// TODO: Add background func started in main that periodically
// iterates over all peers and tries to reconnect to dead ones
// or deletes ones that are dead for longer than 90 minutes
func (p *Peer) restartConnection() error {
	p.conn.Close()
	conn, err := net.Dial(connType, p.addr.ToString())
	if err != nil {
		//log.Error().Err(err).Str("peer", p.addr).Msg("Failed restaring connection to peer")
		p.alive = false
		return err
	}
	p.conn = conn
	return nil
}

func (p Peer) Start() {
	//go p.inConnHandler()
	go p.inHandler()
	go p.outHandler()
}

// inHandler sends messages to other peers
func (p *Peer) outHandler() {
	for range time.Tick(3 * time.Second) {
		if !p.inbound {
			if err := p.WriteMessage(messages.NewPingMessage()); err != nil {
				log.Error("Failed to send ping", "err", err)
			}
			// 30% chance to ask for peers
			if rand.Intn(10) > 7 {
				if err := p.WriteMessage(messages.NewGetAddrMessage()); err != nil {
					p.logger.Error("Failed to send get addr", "err", err)
				}

			}
		}
	}
	//	for range time.Tick(3 * time.Second) {
	//		encoder := gob.NewEncoder(p.conn)
	//		tx := randomTransaction()
	//		log.Debug().Msgf("Sending: %+v", *tx)
	//
	//		if err := encoder.Encode(*tx); err != nil {
	//			log.Error().Err(err).Msg("Failed to send message")
	//			// TODO: Might need to a mutex to the conn for the restarting
	//			if err := p.restartConnection(); err != nil {
	//				return
	//			}
	//		}
	//	}
}

//// inConnHandler takes care of incomming messages i.e.
//// parses then and sends it via channel to the inHandler
//// the separation is put in place to potentially launch multiple
//// inHandlers to increase concurrency
//func (p *Peer) inConnHandler() {
//	for {
//		var header messageHeader
//		// TODO: Extract the decoder as part of the Peer struct (?)
//		decoder := gob.NewDecoder(p.conn)
//		if err := decoder.Decode(&header); err != nil {
//			// TODO: Switch on error type
//			log.Debug().Msgf("Error type: %T", err)
//			log.Error().Err(err).Msg("Received malformed request")
//			// TODO: Kill all goroutines?
//			if err := p.restartConnection(); err != nil {
//				return
//			}
//		}
//		switch header.Command {
//
//		case CmdVersion:
//			version := new(VersionMessage)
//			decoder.Decode(version)
//			if err := decoder.Decode(&header); err != nil {
//				log.Debug().Msgf("Error type: %T", err)
//				log.Error().Err(err).Msg("Received unknown message")
//			}
//			p.inChan <- version
//
//		case CmdVerAck:
//		case CmdGetAddr:
//		case CmdAddr:
//		case CmdTx:
//
//		}
//	}
//}

// inHandler handles executing appropriate actions for incomming messages
// TODO: Switch statement on the current action from channel sent from here in new func
// TODO: Parse type of message from peer (new client, tx, new block, etc.)
// and send into the appropriate channel
func (p *Peer) inHandler() {
	for {
		msg, err := p.ReadMessage()
		if err != nil {
			p.logger.Error("Received malformed request", "err", err)
			if err := p.restartConnection(); err != nil {
				// TODO: Kill all goroutines?
				return
			}
			continue
		}
		switch msg.Command() {
		case messages.CmdVersion:
			p.logger.Warn("Ignoring VersionMessage after initialization")
		case messages.CmdVerAck:
			p.logger.Warn("Ignoring VerAckMessage after initialization")
		case messages.CmdPing:
			p.WriteMessage(messages.NewPongMessage())
			p.logger.Debug("Got ping")
		case messages.CmdPong:
			p.logger.Debug("Sent pong")
		case messages.CmdGetAddr:
			p.logger.Debug("Got Get Address command")
			if err := p.HandleGetAddressMessage(); err != nil {
				p.logger.Error("Failed to send GetAddress", "err", err)
			}
		case messages.CmdAddr:
			p.logger.Debug("Got Address command")
			p.HandleAddressMessage(msg.(messages.AddrMessage))
		default:
			p.logger.Warn("Unknown command")
		}
	}
}

func NewPeer(
	logger log.Logger,
	getPeersCallback func(crypto.FixedHash) []string,
	newOutboundPeerCallback func(string) error,
) Peer {
	return Peer{
		logger:                  logger,
		getPeers:                getPeersCallback,
		newOutboundPeerCallback: newOutboundPeerCallback,
	}
}
