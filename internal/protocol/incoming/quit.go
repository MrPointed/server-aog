package incoming

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type QuitPacket struct {
}

func (p *QuitPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	connection.Send(&outgoing.DisconnectPacket{})
	connection.Disconnect()
	return true, nil
}
