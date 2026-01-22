package outgoing

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/model"
)

type NpcCreatePacket struct {
	Npc *model.WorldNPC
}

func (p *NpcCreatePacket) Write(buffer *network.DataBuffer) error {
	if p.Npc == nil || p.Npc.NPC == nil {
		buffer.PutShort(0)
		buffer.PutShort(0)
		buffer.PutShort(0)
		buffer.Put(1)
		buffer.Put(1)
		buffer.Put(1)
		buffer.PutShort(0)
		buffer.PutShort(0)
		buffer.PutShort(0)
		buffer.PutShort(0)
		buffer.PutShort(0)
		buffer.PutUTF8String("Unknown")
		buffer.Put(0)
		buffer.Put(0)
		return nil
	}

	buffer.PutShort(p.Npc.Index)
	buffer.PutShort(int16(p.Npc.NPC.Body))
	buffer.PutShort(int16(p.Npc.NPC.Head))
	buffer.Put(byte(p.Npc.Heading + 1))
	buffer.Put(p.Npc.Position.X + 1)
	buffer.Put(p.Npc.Position.Y + 1)
	
	buffer.PutShort(0) // Weapon
	buffer.PutShort(0) // Shield
	buffer.PutShort(0) // Helmet
	
	buffer.PutShort(0) // Fx ID
	buffer.PutShort(0) // Fx Loops
	
	buffer.PutUTF8String(p.Npc.NPC.Name)
	
	// Nick Color
	if p.Npc.NPC.Hostile {
		buffer.Put(0x01) // Criminal
	} else {
		buffer.Put(0x02) // Attackable/Friendly
	}
	
	buffer.Put(0) // Privileges
	
	return nil
}
