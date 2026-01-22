package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type ChangeMapPacket struct {
	MapId   int
	Version int16
}

func (p *ChangeMapPacket) Write(buffer *network.DataBuffer) error {
	buffer.PutShort(int16(p.MapId))
	buffer.PutShort(p.Version)
	return nil
}
