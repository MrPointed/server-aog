package incoming

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
)

type RequestPositionUpdatePacket struct {
}

func (p *RequestPositionUpdatePacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	// Consumes no extra bytes
	return true, nil
}
