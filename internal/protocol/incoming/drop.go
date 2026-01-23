package incoming

import (
	"fmt"

	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/service"
)

type DropPacket struct {
	MapService     *service.MapService
	MessageService *service.MessageService
	ObjectService  *service.ObjectService
}

func (p *DropPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	if buffer.ReadableBytes() < 3 {
		return false, nil
	}

	slotIdx, _ := buffer.Get()
	amountShort, _ := buffer.GetShort()

	slot := int(slotIdx) - 1
	amount := int(amountShort)

	char := connection.GetUser()
	if char == nil {
		return true, nil
	}

	if char.Dead {
		return true, nil
	}

	itemSlot := char.Inventory.GetSlot(slot)
	if itemSlot == nil || itemSlot.ObjectID == 0 || itemSlot.Amount < amount {
		return true, nil
	}

	if itemSlot.Equipped {
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "No puedes tirar un objeto equipado.",
			Font:    outgoing.INFO,
		})
		return true, nil
	}

	obj := p.ObjectService.GetObject(itemSlot.ObjectID)
	if obj == nil {
		return true, nil
	}

	// Check if map tile is empty of objects
	if p.MapService.GetObjectAt(char.Position) != nil {
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "Ya hay un objeto en esta posición.",
			Font:    outgoing.INFO,
		})
		return true, nil
	}

	// Drop logic
	itemSlot.Amount -= amount
	dropAmount := amount
	if itemSlot.Amount <= 0 {
		itemSlot.ObjectID = 0
		itemSlot.Amount = 0
	}

	// Update map
	worldObj := &model.WorldObject{
		Object: obj,
		Amount: dropAmount,
	}
	p.MapService.PutObject(char.Position, worldObj)

	// Broadcast appearance
	p.MessageService.SendToArea(&outgoing.ObjectCreatePacket{
		X:            char.Position.X,
		Y:            char.Position.Y,
		GraphicIndex: int16(obj.GraphicIndex),
	}, char.Position)

	// Sync inventory
	connection.Send(&outgoing.ChangeInventorySlotPacket{
		Slot:     slotIdx,
		Object:   p.ObjectService.GetObject(itemSlot.ObjectID),
		Amount:   itemSlot.Amount,
		Equipped: itemSlot.Equipped,
	})

	connection.Send(&outgoing.ConsoleMessagePacket{
		Message: fmt.Sprintf("Has tirado %d %s.", dropAmount, obj.Name),
		Font:    outgoing.INFO,
	})

	return true, nil
}
