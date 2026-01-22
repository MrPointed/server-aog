package incoming

import (
	"fmt"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/service"
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/utils"
)

type UseItemPacket struct {
	ObjectService  *service.ObjectService
	MessageService *service.MessageService
}

func (p *UseItemPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	if buffer.ReadableBytes() < 1 {
		return false, nil
	}

	slotIdx, _ := buffer.Get()
	slot := int(slotIdx) - 1

	char := connection.GetUser()
	if char == nil {
		return true, nil
	}

	itemSlot := char.Inventory.GetSlot(slot)
	if itemSlot == nil || itemSlot.ObjectID == 0 {
		return true, nil
	}

	obj := p.ObjectService.GetObject(itemSlot.ObjectID)
	if obj == nil {
		return true, nil
	}

	switch obj.Type {
	case model.OTFood:
		char.Hunger = utils.Min(100, char.Hunger + obj.HungerPoints)
		p.MessageService.SendToArea(outgoing.NewUpdateUserStatsPacket(char), char.Position)
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: fmt.Sprintf("Has comido %s.", obj.Name),
			Font:    outgoing.INFO,
		})
		// Consumption logic (remove 1 item)
		itemSlot.Amount--
		if itemSlot.Amount <= 0 {
			itemSlot.ObjectID = 0
		}
		// Sync inventory
		connection.Send(&outgoing.ChangeInventorySlotPacket{
			Slot: slotIdx,
			Object: p.ObjectService.GetObject(itemSlot.ObjectID),
			Amount: itemSlot.Amount,
			Equipped: itemSlot.Equipped,
		})

	case model.OTDrink:
		char.Thirstiness = utils.Min(100, char.Thirstiness + obj.ThirstPoints)
		p.MessageService.SendToArea(outgoing.NewUpdateUserStatsPacket(char), char.Position)
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: fmt.Sprintf("Has bebido %s.", obj.Name),
			Font:    outgoing.INFO,
		})
		itemSlot.Amount--
		if itemSlot.Amount <= 0 {
			itemSlot.ObjectID = 0
		}
		connection.Send(&outgoing.ChangeInventorySlotPacket{
			Slot: slotIdx,
			Object: p.ObjectService.GetObject(itemSlot.ObjectID),
			Amount: itemSlot.Amount,
			Equipped: itemSlot.Equipped,
		})

	case model.OTPotion:
		// TODO: Implement potion effects (HP, Mana, etc)
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "Las pociones no están implementadas aún.",
			Font:    outgoing.INFO,
		})

	default:
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "No puedes usar este objeto.",
			Font:    outgoing.INFO,
		})
	}

	return true, nil
}