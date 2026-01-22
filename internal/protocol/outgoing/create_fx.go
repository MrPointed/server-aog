package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type CreateFxPacket struct {
	CharIndex int16
	FxID      int16
	Loops     int16
}

func (p *CreateFxPacket) Write(buffer *network.DataBuffer) error {
	buffer.PutShort(p.CharIndex)
	buffer.PutShort(p.FxID)
	buffer.PutShort(p.Loops)
	return nil
}
