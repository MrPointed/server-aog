package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type DiceRollPacket struct {
	Strength     byte
	Dexterity    byte
	Intelligence byte
	Charisma     byte
	Constitution byte
}

func (p *DiceRollPacket) Write(buffer *network.DataBuffer) error {
	buffer.Put(p.Strength)
	buffer.Put(p.Dexterity)
	buffer.Put(p.Intelligence)
	buffer.Put(p.Charisma)
	buffer.Put(p.Constitution)
	return nil
}
