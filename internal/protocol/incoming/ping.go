package incoming

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type PingPacket struct {
}

func (p *PingPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	connection.Send(&outgoing.PongPacket{})
	return true, nil
}
