package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type ObjectCreatePacket struct {
	X            byte
	Y            byte
	GraphicIndex int16
}

func (p *ObjectCreatePacket) Write(buffer *network.DataBuffer) error {
	buffer.Put(p.X + 1)
	buffer.Put(p.Y + 1)
	buffer.PutShort(p.GraphicIndex)
	return nil
}
