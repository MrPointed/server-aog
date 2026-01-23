package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type CommerceEndPacket struct {
}

func (p *CommerceEndPacket) Write(buffer *network.DataBuffer) error {
	return nil
}
