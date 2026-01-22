package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type CharacterMovePacket struct {
	CharIndex int16
	X         byte
	Y         byte
}

func (p *CharacterMovePacket) Write(buffer *network.DataBuffer) error {
	buffer.PutShort(p.CharIndex)
	buffer.Put(p.X + 1)
	buffer.Put(p.Y + 1)
	return nil
}
