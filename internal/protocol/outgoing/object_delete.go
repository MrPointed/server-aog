package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type ObjectDeletePacket struct {
	X byte
	Y byte
}

func (p *ObjectDeletePacket) Write(buffer *network.DataBuffer) error {
	buffer.Put(p.X + 1)
	buffer.Put(p.Y + 1)
	return nil
}
