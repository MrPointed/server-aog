package outgoing

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/model"
)

type CharacterCreatePacket struct {
	Character *model.Character
}

func (p *CharacterCreatePacket) Write(buffer *network.DataBuffer) error {
	buffer.PutShort(p.Character.CharIndex)
	buffer.PutShort(int16(p.Character.Body))
	buffer.PutShort(int16(p.Character.Head))
	buffer.Put(byte(p.Character.Heading + 1))
	buffer.Put(p.Character.Position.X + 1)
	buffer.Put(p.Character.Position.Y + 1)
	
	buffer.PutShort(p.Character.Weapon)
	buffer.PutShort(p.Character.Shield)
	buffer.PutShort(p.Character.Helmet)
	
	buffer.PutShort(0) // Fx ID
	buffer.PutShort(0) // Fx Loops
	
	buffer.PutUTF8String(p.Character.Name) // Java uses putUnicodeString which is UTF-8 with 2 byte len
	buffer.Put(0) // Nick Color
	buffer.Put(byte(p.Character.Privileges)) // Privileges Flags
	
	return nil
}
