package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type BlockPositionPacket struct {
	X       byte
	Y       byte
	Blocked bool
}

func (p *BlockPositionPacket) Write(buffer *network.DataBuffer) error {
	buffer.Put(p.X + 1)
	buffer.Put(p.Y + 1)
	if p.Blocked {
		buffer.Put(1)
	} else {
		buffer.Put(0)
	}
	return nil
}
