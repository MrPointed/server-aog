package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type ChangeSpellSlotPacket struct {
	Slot      byte
	SpellID   int16
	SpellName string
}

func (p *ChangeSpellSlotPacket) Write(buffer *network.DataBuffer) error {
	buffer.Put(p.Slot)
	buffer.PutShort(p.SpellID)
	buffer.PutUTF8String(p.SpellName)
	return nil
}
