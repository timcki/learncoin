package messages

import "github.com/timcki/learncoin/internal/crypto"

const (
	// Used to identify new nodes in the network
	CmdVersion = "version"
	CmdVerAck  = "verack"

	// Used to get the addresses of neighboring nodes
	CmdGetAddr = "getaddr"
	CmdAddr    = "addr"
	CmdPing    = "ping"
	CmdPong    = "pong"

	// Placeholder to exchange some data between nodes when connected
	CmdTx = "tx"
)

type VersionMessage struct {
	Version string
	Address string
	ID      crypto.FixedHash
	Nonce   uint64 // Random nonce to detect connections to self
}

func NewVersionMessage(version, addr string, id crypto.FixedHash) Message {
	return &VersionMessage{
		Version: version,
		Address: addr,
		ID:      id,
		Nonce:   0,
	}
}

func (v VersionMessage) Command() string {
	return CmdVersion
}

type VerAckMessage struct{}

func (v VerAckMessage) Command() string {
	return CmdVerAck
}

func NewVerAckMessage() Message {
	return new(VerAckMessage)
}

type GetAddrMessage struct{}

func NewGetAddrMessage() Message {
	return &GetAddrMessage{}
}

func (m GetAddrMessage) Command() string {
	return CmdGetAddr
}

type AddrMessage struct {
	Nodes []string
}

func NewAddrMessage(nodes []string) Message {
	return &AddrMessage{
		Nodes: nodes,
	}
}

func (m AddrMessage) Command() string {
	return CmdAddr
}

type PingMessage struct {
	Ping string
}
type PongMessage struct {
	Pong string
}

func (m PingMessage) Command() string {
	return CmdPing
}

func (m PongMessage) Command() string {
	return CmdPong
}

func NewPingMessage() Message {
	return new(PingMessage)
}

func NewPongMessage() Message {
	return new(PongMessage)
}

type Msg interface {
	MessageHeader | PingMessage | PongMessage | AddrMessage | GetAddrMessage | VerAckMessage | VersionMessage
}

// Interface that struct must implement to
// be a valid message in the learncoin protocol
type Message interface {
	Command() string
}

// messageHeader defines the header structure of all
// messages in the learncoin protocol. It has a fixed
// size which simplifies marshalling/unmarshalling
// Size: 14 bytes
type MessageHeader struct {
	Cmd string // 10 bytes (fixed)
	// length   uint32  // 4 bytes
	// checksum [4]byte // 4 bytes
}

func (h MessageHeader) Command() string {
	return h.Cmd
}
