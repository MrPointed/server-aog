package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type UserCharIndexInServerPacket struct {
	UserIndex int16
}

func (p *UserCharIndexInServerPacket) Write(buffer *network.DataBuffer) error {
	buffer.PutShort(p.UserIndex)
	return nil
}

type UserIndexInServerPacket struct {
	UserIndex  int16
}

func (p *UserIndexInServerPacket) Write(buffer *network.DataBuffer) error {
	buffer.PutShort(p.UserIndex)
	return nil
}
