package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type PlayWavePacket struct {
	Wave byte
	X    byte
	Y    byte
}

func (p *PlayWavePacket) Write(buffer *network.DataBuffer) error {
	buffer.Put(p.Wave)
	buffer.Put(p.X + 1)
	buffer.Put(p.Y + 1)
	return nil
}
