package incoming

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/service"
)

type EquipItemPacket struct {
	ItemActionService service.ItemActionService
}

func (p *EquipItemPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	if buffer.ReadableBytes() < 1 {
		return false, nil
	}

	slotIdx, _ := buffer.Get()
	slot := int(slotIdx) - 1

	char := connection.GetUser()
	if char == nil {
		return true, nil
	}

	p.ItemActionService.EquipItem(char, slot, connection)

	return true, nil
}