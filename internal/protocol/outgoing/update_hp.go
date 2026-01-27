package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type UpdateHPPacket struct {
	Hp int
}

func (p *UpdateHPPacket) Write(buffer *network.DataBuffer) error {
	buffer.Put(0) // Dummy byte
	buffer.PutShort(int16(p.Hp))
	return nil
}
