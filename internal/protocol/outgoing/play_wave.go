package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type PlayWavePacket struct {
	Wave int16
	X    byte
	Y    byte
}

func (p *PlayWavePacket) Write(buffer *network.DataBuffer) error {
	buffer.PutShort(p.Wave)
	buffer.Put(p.X + 1)
	buffer.Put(p.Y + 1)
	return nil
}
