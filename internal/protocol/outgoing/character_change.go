package outgoing

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/model"
)

type CharacterChangePacket struct {
	Character *model.Character
}

func (p *CharacterChangePacket) Write(buffer *network.DataBuffer) error {
	buffer.PutShort(p.Character.CharIndex)
	buffer.PutShort(int16(p.Character.Body))
	buffer.PutShort(int16(p.Character.Head))
	buffer.Put(byte(p.Character.Heading + 1))
	
	buffer.PutShort(p.Character.Weapon)
	buffer.PutShort(p.Character.Shield)
	buffer.PutShort(p.Character.Helmet)
	
	buffer.PutShort(0) // Fx ID
	buffer.PutShort(0) // Fx Loops
	
	return nil
}
