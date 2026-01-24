package outgoing

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/model"
)

type CharacterMovePacket struct {
	CharIndex int16
	X         byte
	Y         byte
	Heading   model.Heading
}

func (p *CharacterMovePacket) Write(buffer *network.DataBuffer) error {
	buffer.PutShort(p.CharIndex)
	buffer.Put(p.X + 1)
	buffer.Put(p.Y + 1)
	// We don't add heading here yet because we need to check if the client expects it in packet 32.
	// Traditional AO packet 32 is [SHORT Index, BYTE X, BYTE Y]. 
	// If we want to change it, we must change the client too.
	return nil
}
