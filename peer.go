package main

import (
	"encoding/gob"
	"errors"
	"net"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/timcki/learncoin/messages"
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

	// Safe for concurrent access, set at creation and won't change
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

// Connects to a new peer (sends CmdVersion and waits for CmdVerAck)
func NewOutboundPeer(address string) (peer *Peer, err error) {
	peer = new(Peer)
	peer.conn, err = net.Dial(connType, address)
	if err != nil {
		log.Error().Err(err).Str("peer", address).Msg("Failed connection to peer")
		return
	}
	if err = peer.sendVersionMessage(); err != nil {
		return
	}
	if err = peer.handleVersionMessage(); err != nil {
		return
	}
	peer.inbound = false
	peer.alive = true
	log.Debug().Str("peer", peer.addr).Msg("Succesfully registered outbound peer")
	return
}

// NewInboundPeer handles the connection of a new peer
func NewInboundPeer(conn net.Conn) (peer *Peer, err error) {
	peer = new(Peer)
	peer.conn = conn
	if err = peer.handleVersionMessage(); err != nil {
		log.Error().Err(err).Msg("Failed connection to peer")
		return
	}
	if err = peer.sendVersionMessage(); err != nil {
		log.Error().Err(err).Msg("Failed send to peer")
		return
	}
	peer.inbound = true
	peer.alive = true
	log.Debug().Str("peer", peer.addr).Msg("Succesfully registered inbound peer")
	return
}

func (p *Peer) readMessage() (message messages.Message, err error) {
	decoder := gob.NewDecoder(p.conn)
	if err = decoder.Decode(&message); err != nil {
		//log.Error().Err(err).Msg("Failed to parse message")
		return
	}
	return
}

func (p *Peer) writeMessage(msg messages.Message) error {
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
func (p *Peer) sendVersionMessage() error {
	msg := messages.NewVersionMessage(connHost, connPort)
	if err := p.writeMessage(msg); err != nil {
		return err
	}
	verack, err := p.readMessage()
	if err != nil || verack.Command() != messages.CmdVerAck {
		//log.Error().Err(err).Str("peer", p.addr).Msg("Didn't receive VerAckMessage")
		return err
	}
	return nil

}

// handleVersionMessage reads a message from the peer network connection
// and parses it if it's a VersionMessage and sends a VerAck, otherwise returns error
func (p *Peer) handleVersionMessage() error {
	msg, err := p.readMessage()
	if err != nil {
		return err
	}
	if msg.Command() == messages.CmdVersion {
		if err := p.writeMessage(messages.NewVerAckMessage()); err != nil {
			return err
		}
		p.addr = msg.(messages.VersionMessage).Address
		log.Debug().Msgf("%s", p.addr)

		return nil
	}
	return NoVersionMessageOnInitError
}

// TODO: Add background func started in main that periodically
// iterates over all peers and tries to reconnect to dead ones
// or deletes ones that are dead for longer than 90 minutes
func (p *Peer) restartConnection() error {
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

func (p *Peer) start() error {
	//go p.inConnHandler()
	go p.inHandler()
	go p.outHandler()
	return nil
}

// inHandler sends messages to other peers
func (p *Peer) outHandler() {
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
func (p *Peer) inHandler() {
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
