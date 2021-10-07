package messages

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
	Address string
	Nonce   uint64 // Random nonce to detect connections to self
}

func NewVersionMessage(addr, port string) Message {
	return &VersionMessage{
		Address: addr + ":" + port,
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
	return new(GetAddrMessage)
}

func (m GetAddrMessage) Command() string {
	return CmdGetAddr
}

type AddrMessage struct {
	Count int
	Nodes []string
}

func NewAddrMessage(nodes []string) Message {
	return &AddrMessage{
		Count: len(nodes),
		Nodes: nodes,
	}
}

func (m AddrMessage) Command() string {
	return CmdAddr
}

type PingMessage struct{}
type PongMessage struct{}

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

// Interface that struct must implement to
// be a valid message in the learncoin protocol
type Message interface {
	Command() string
}

// messageHeader defines the header structure of all
// messages in the learncoin protocol. It has a fixed
// size which simplifies marshalling/unmarshalling
// Size: 14 bytes
// TODO: Implement a checksum
// NOTE: Currently unused
type messageHeader struct {
	Command string // 10 bytes (fixed)
	// length   uint32  // 4 bytes
	// checksum [4]byte // 4 bytes
}
