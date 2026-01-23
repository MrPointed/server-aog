package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type DisconnectPacket struct {
}

func (p *DisconnectPacket) Write(buffer *network.DataBuffer) error {
	return nil
}
