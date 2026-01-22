package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type CharacterRemovePacket struct {
	CharIndex int16
}

func (p *CharacterRemovePacket) Write(buffer *network.DataBuffer) error {
	buffer.PutShort(p.CharIndex)
	return nil
}
