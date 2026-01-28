package outgoing

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/model"
)

type ChangeNpcInventorySlotPacket struct {
	Slot   byte
	Object *model.Object
	Amount int
}

func (p *ChangeNpcInventorySlotPacket) Write(buffer *network.DataBuffer) error {
	buffer.Put(p.Slot)
	
	if p.Object == nil {
		buffer.PutUTF8String("")
		buffer.PutShort(0)
		buffer.PutFloat(0)
		buffer.PutShort(0)
		buffer.PutShort(0)
		buffer.Put(0)
		buffer.PutShort(0)
		buffer.PutShort(0)
		buffer.PutShort(0)
		buffer.PutShort(0)
		return nil
	}

	buffer.PutUTF8String(p.Object.Name)
	buffer.PutShort(int16(p.Amount))
	buffer.PutFloat(float32(p.Object.Value))
	buffer.PutShort(int16(p.Object.GraphicIndex))
	buffer.PutShort(int16(p.Object.ID))
	buffer.Put(byte(p.Object.Type))
	buffer.PutShort(int16(p.Object.MaxHit))
	buffer.PutShort(int16(p.Object.MinHit))
	buffer.PutShort(int16(p.Object.MaxDef))
	buffer.PutShort(int16(p.Object.MinDef))
	
	return nil
}
