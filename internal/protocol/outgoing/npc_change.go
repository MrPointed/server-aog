package outgoing

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/model"
)

type NpcChangePacket struct {
	Npc *model.WorldNPC
}

func (p *NpcChangePacket) Write(buffer *network.DataBuffer) error {
	buffer.PutShort(p.Npc.Index)
	buffer.PutShort(int16(p.Npc.NPC.Body))
	buffer.PutShort(int16(p.Npc.NPC.Head))
	buffer.Put(byte(p.Npc.Heading + 1))
	
	buffer.PutShort(0) // Weapon
	buffer.PutShort(0) // Shield
	buffer.PutShort(0) // Helmet
	
	buffer.PutShort(0) // Fx ID
	buffer.PutShort(0) // Fx Loops
	
	return nil
}
