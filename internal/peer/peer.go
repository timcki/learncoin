package peer

import (
	"encoding/gob"
	"errors"
	"net"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

type Peer interface {
	SetConn(conn net.Conn)
	SetInbound(b bool)
	SetAlive(b bool)
	SetAddr(addr string)

	GetConn() net.Conn
	GetAddr() string
	// TODO: Create type for node id
	GetID() crypto.Hash
	IsInboud() bool
	IsAlive() bool

	SendVersionMessage(messages.VersionMessage) error
	HandleVersionMessage() error

	Start()
}

type peer struct {
	// Connection to the peer. From doc:
	// Multiple goroutines may invoke methods on a Conn simultaneously.
	// So no mutex needed
	conn net.Conn

	// Logger dedicated to the peer
	logger zerolog.Logger

	// Safe for concurrent access, set at creation and won't change
	id      crypto.Hash
	addr    string
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
}

func (p peer) SetConn(conn net.Conn) {
	p.conn = conn
}

func (p peer) SetInbound(b bool) {
	p.inbound = b
}

func (p peer) SetAlive(b bool) {
	p.alive = b
}

func (p peer) SetAddr(addr string) {
	p.addr = addr
}

func (p peer) GetConn() net.Conn {
	return p.conn
}

func (p peer) GetAddr() string {
	return p.addr
}

func (p peer) GetID() crypto.Hash {
	return p.id
}

func (p peer) IsInboud() bool {
	return p.inbound
}
func (p peer) IsAlive() bool {
	return p.alive
}

func (p *peer) readMessage() (message messages.Message, err error) {
	decoder := gob.NewDecoder(p.GetConn())
	if err = decoder.Decode(&message); err != nil {
		//log.Error().Err(err).Msg("Failed to parse message")
		return
	}
	return
}

func (p *peer) writeMessage(msg messages.Message) error {
	// TODO: (?) Create global encoder and hold it in structure
	encoder := gob.NewEncoder(p.conn)
	if err := encoder.Encode(&msg); err != nil {
		//log.Error().Err(err).Str("peer", p.addr).Msgf("Failed to send message %s", msg.Command())
		return err
	}
	return nil

}

// sendVersionMessage sends a VersionMessage and waits for VerAck
// otherwise fails
func (p peer) SendVersionMessage(msg messages.VersionMessage) error {
	//msg := messages.NewVersionMessage(connHost, p.id)
	if err := p.writeMessage(msg); err != nil {
		return err
	}
	verack, err := p.readMessage()
	if err != nil || verack.Command() != messages.CmdVerAck {
		p.logger.Error().Err(err).Str("peer", p.addr).Msg("Didn't receive VerAckMessage")
		return err
	}
	return nil
}

// handleVersionMessage reads a message from the peer network connection,
// parses it if it's a VersionMessage and sends a VerAck, otherwise returns error
func (p peer) HandleVersionMessage() error {
	msg, err := p.readMessage()
	if err != nil {
		return err
	}
	if msg.Command() == messages.CmdVersion {
		// Receive VersionMessage and send verack before actually registering the peer
		if err := p.writeMessage(messages.NewVerAckMessage()); err != nil {
			return err
		}
		// Fill info from VersionMessage
		p.addr = msg.(messages.VersionMessage).Address
		p.id = msg.(messages.VersionMessage).ID
		p.logger.Debug().Msgf("%s", p.addr)

		return nil
	}
	return NoVersionMessageOnInitError
}

// TODO: Add background func started in main that periodically
// iterates over all peers and tries to reconnect to dead ones
// or deletes ones that are dead for longer than 90 minutes
func (p *peer) restartConnection() error {
	p.conn.Close()
	conn, err := net.Dial(connType, p.addr)
	if err != nil {
		//log.Error().Err(err).Str("peer", p.addr).Msg("Failed restaring connection to peer")
		p.alive = false
		return err
	}
	p.conn = conn
	return nil
}

func (p peer) Start() {
	//go p.inConnHandler()
	go p.inHandler()
	go p.outHandler()
}

// inHandler sends messages to other peers
func (p *peer) outHandler() {
	for range time.Tick(3 * time.Second) {
		if err := p.writeMessage(messages.NewPingMessage()); err != nil {
			log.Error().Err(err).Msg("Failed to send ping")
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
func (p *peer) inHandler() {
	for {
		msg, err := p.readMessage()
		if err != nil {
			log.Error().Err(err).Str("peer", p.addr).Msg("Received malformed request")
			if err := p.restartConnection(); err != nil {
				// TODO: Kill all goroutines?
				return
			}
			continue
		}
		switch msg.Command() {
		case messages.CmdVersion:
			log.Warn().Str("peer", p.addr).Msg("Ignoring VersionMessage after initialization")
		case messages.CmdVerAck:
			log.Warn().Str("peer", p.addr).Msg("Ignoring VerAckMessage after initialization")
		case messages.CmdPing:
			p.writeMessage(messages.NewPongMessage())
			log.Debug().Str("peer", p.addr).Msg("Sent Ping")
		case messages.CmdPong:
			log.Debug().Str("peer", p.addr).Msg("Got Pong")
		case messages.CmdGetAddr:
			log.Debug().Str("peer", p.addr).Msg("Got CmdGetAddr")
		default:
			log.Error().Msg("Unknown command")

		}
		//		var tx Transaction
		//		decoder := gob.NewDecoder(p.conn)
		//		err := decoder.Decode(&tx)
		//		if err != nil {
		//			log.Error().Err(err).Msg("Received malformed request")
		//			p.conn.Close()
		//			return
		//		}
		//		log.Info().Msgf("Got message: %+v from peer %s", tx, p.addr)
	}
}

func NewPeer(logger zerolog.Logger) Peer {
	return peer{
		logger: logger,
	}
}
