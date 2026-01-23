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
	ObjectService   *service.ObjectService
	MessageService  *service.MessageService
	IntervalService *service.IntervalService
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

	// Check intervals
	if !p.IntervalService.CanUseItem(char) {
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
		p.IntervalService.UpdateLastItem(char)

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
		p.IntervalService.UpdateLastItem(char)

	case model.OTPotion:
		switch obj.PotionType {
		case 1: // Agility
			connection.Send(&outgoing.ConsoleMessagePacket{
				Message: "Poción de agilidad no implementada aún.",
				Font:    outgoing.INFO,
			})
		case 2: // Strength
			connection.Send(&outgoing.ConsoleMessagePacket{
				Message: "Poción de fuerza no implementada aún.",
				Font:    outgoing.INFO,
			})
		case 3: // HP
			modifier := utils.RandomNumber(obj.MinModifier, obj.MaxModifier)
			char.Hp = utils.Min(char.MaxHp, char.Hp + modifier)
			connection.Send(&outgoing.ConsoleMessagePacket{
				Message: fmt.Sprintf("Has recuperado %d puntos de vida.", modifier),
				Font:    outgoing.INFO,
			})
		case 4: // Mana
			modifier := utils.RandomNumber(obj.MinModifier, obj.MaxModifier)
			char.Mana = utils.Min(char.MaxMana, char.Mana + modifier)
			connection.Send(&outgoing.ConsoleMessagePacket{
				Message: fmt.Sprintf("Has recuperado %d puntos de mana.", modifier),
				Font:    outgoing.INFO,
			})
		case 5: // Poison
			char.Poisoned = false
			connection.Send(&outgoing.ConsoleMessagePacket{
				Message: "Te has curado del envenenamiento.",
				Font:    outgoing.INFO,
			})
		}
		// Consumption logic
		itemSlot.Amount--
		if itemSlot.Amount <= 0 {
			itemSlot.ObjectID = 0
		}
		// Sync
		p.MessageService.SendToArea(outgoing.NewUpdateUserStatsPacket(char), char.Position)
		connection.Send(&outgoing.ChangeInventorySlotPacket{
			Slot:     slotIdx,
			Object:   p.ObjectService.GetObject(itemSlot.ObjectID),
			Amount:   itemSlot.Amount,
			Equipped: itemSlot.Equipped,
		})
		p.IntervalService.UpdateLastItem(char)

	case model.OTMoney:
		goldAmount := itemSlot.Amount
		char.Gold += goldAmount
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: fmt.Sprintf("Has guardado %d monedas de oro en tu billetera.", goldAmount),
			Font:    outgoing.INFO,
		})
		itemSlot.Amount = 0
		itemSlot.ObjectID = 0
		// Sync
		p.MessageService.SendToArea(outgoing.NewUpdateUserStatsPacket(char), char.Position)
		connection.Send(&outgoing.ChangeInventorySlotPacket{
			Slot:     slotIdx,
			Object:   nil,
			Amount:   0,
			Equipped: false,
		})
		p.IntervalService.UpdateLastItem(char)

	default:
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "No puedes usar este objeto.",
			Font:    outgoing.INFO,
		})
	}

	return true, nil
}