package main

import (
	"encoding/gob"
	"net"
	"time"

	"github.com/rs/zerolog/log"
)

type Peer struct {
	// Connection to the peer
	conn net.Conn

	// Safe for concurrent access, set at creation and won't change
	addr    string
	inbound bool
}

// Creates a new peer from an address
func newOutboundPeer(address string) (peer *Peer, err error) {
	peer = new(Peer)
	conn, err := net.Dial(connType, address)
	if err != nil {
		log.Error().Err(err).Str("peer", address).Msg("Failed connection to peer")
		return
	}
	peer.conn = conn
	peer.addr = address
	peer.inbound = false
	return
}

// No place to fail but I'll leave the error return for future improvements
// and uniformity with newOutboundPeer
func newInboundPeer(conn net.Conn) (*Peer, error) {
	peer := new(Peer)
	peer.conn = conn
	peer.addr = conn.RemoteAddr().String()
	peer.inbound = true
	return peer, nil
}

func (p *Peer) start() error {
	go p.inHandler()
	go p.outHandler()
	return nil
}

// inHandler sends messages to other peers
func (p *Peer) inHandler() {
	for range time.Tick(3 * time.Second) {
		encoder := gob.NewEncoder(p.conn)
		tx := randomTransaction()
		log.Debug().Msgf("Sending: %+v", *tx)

		if err := encoder.Encode(*tx); err != nil {
			log.Error().Err(err).Msg("Failed to send message")
			return
		}
	}
}

// outHandler handles incomming communication.
// TODO: Switch statement on the current action from channel sent from here in new func
// TODO: Parse type of message from peer (new client, tx, new block, etc.) and send into the appropriate channel
func (p *Peer) outHandler() {
	for {
		var tx Transaction
		decoder := gob.NewDecoder(p.conn)
		err := decoder.Decode(&tx)
		if err != nil {
			log.Error().Err(err).Msg("Received malformed request")
			p.conn.Close()
			return
		}
		log.Info().Msgf("Got message: %+v from peer %s", tx, p.addr)
	}
}
