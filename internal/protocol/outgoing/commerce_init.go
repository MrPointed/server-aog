package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type CommerceInitPacket struct {
}

func (p *CommerceInitPacket) Write(buffer *network.DataBuffer) error {
	// No extra data needed, just the packet ID handled by manager
	return nil
}
