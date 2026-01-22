package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type UpdateStrengthAndDexterityPacket struct {
	Strength  byte
	Dexterity byte
}

func (p *UpdateStrengthAndDexterityPacket) Write(buffer *network.DataBuffer) error {
	buffer.Put(p.Strength)
	buffer.Put(p.Dexterity)
	return nil
}
