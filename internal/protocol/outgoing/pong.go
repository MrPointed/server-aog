package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type PongPacket struct {
}

func (p *PongPacket) Write(buffer *network.DataBuffer) error {
	return nil
}
