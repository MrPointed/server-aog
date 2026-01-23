package incoming

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type RequestAttributesPacket struct {
}

func (p *RequestAttributesPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	char := connection.GetUser()
	if char == nil {
		return true, nil
	}

	connection.Send(&outgoing.AttributesPacket{
		Attributes: char.Attributes,
	})

	return true, nil
}
