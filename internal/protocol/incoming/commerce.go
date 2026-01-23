package incoming

import (
	"fmt"
	"math"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/service"
	"github.com/ao-go-server/internal/model"
)

type CommerceBuyPacket struct {
	NpcService    *service.NpcService
	ObjectService *service.ObjectService
	MessageService *service.MessageService
}

func (p *CommerceBuyPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	if buffer.ReadableBytes() < 3 { // byte slot + short amount
		return false, nil
	}

	slotIdx, _ := buffer.Get()
	amount, _ := buffer.GetShort()

	user := connection.GetUser()
	if user == nil || user.TradingNPCIndex == 0 {
		return true, nil
	}

	npc := p.NpcService.GetWorldNpcByIndex(user.TradingNPCIndex)
	if npc == nil {
		user.TradingNPCIndex = 0
		return true, nil
	}

	// Range check
	dist := int(math.Abs(float64(user.Position.X)-float64(npc.Position.X)) + math.Abs(float64(user.Position.Y)-float64(npc.Position.Y)))
	if dist > 5 {
		connection.Send(&outgoing.ConsoleMessagePacket{Message: "Estás demasiado lejos del vendedor.", Font: outgoing.INFO})
		return true, nil
	}

	slot := int(slotIdx) - 1
	if slot < 0 || slot >= len(npc.NPC.Inventory) {
		return true, nil
	}

	npcSlot := npc.NPC.Inventory[slot]
	obj := p.ObjectService.GetObject(npcSlot.ObjectID)
	if obj == nil {
		return true, nil
	}

	totalPrice := obj.Value * int(amount)
	if user.Gold < totalPrice {
		connection.Send(&outgoing.ConsoleMessagePacket{Message: "No tienes suficiente oro.", Font: outgoing.INFO})
		return true, nil
	}

	if user.Inventory.AddItem(obj.ID, int(amount)) {
		user.Gold -= totalPrice
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: fmt.Sprintf("Has comprado %d %s por %d monedas de oro.", amount, obj.Name, totalPrice),
			Font:    outgoing.INFO,
		})
		
		// Update user stats (gold)
		connection.Send(outgoing.NewUpdateUserStatsPacket(user))
		
		// Sync inventory
		for i := 0; i < model.InventorySlots; i++ {
			if user.Inventory.Slots[i].ObjectID == obj.ID {
				connection.Send(&outgoing.ChangeInventorySlotPacket{
					Slot:     byte(i + 1),
					Object:   obj,
					Amount:   user.Inventory.Slots[i].Amount,
					Equipped: user.Inventory.Slots[i].Equipped,
				})
			}
		}
	} else {
		connection.Send(&outgoing.ConsoleMessagePacket{Message: "No tienes espacio en el inventario.", Font: outgoing.INFO})
	}

	return true, nil
}

type CommerceSellPacket struct {
	NpcService    *service.NpcService
	ObjectService *service.ObjectService
	MessageService *service.MessageService
}

func (p *CommerceSellPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	if buffer.ReadableBytes() < 3 {
		return false, nil
	}

	slotIdx, _ := buffer.Get()
	amount, _ := buffer.GetShort()

	user := connection.GetUser()
	if user == nil || user.TradingNPCIndex == 0 {
		return true, nil
	}

	npc := p.NpcService.GetWorldNpcByIndex(user.TradingNPCIndex)
	if npc == nil {
		user.TradingNPCIndex = 0
		return true, nil
	}

	slot := int(slotIdx) - 1
	itemSlot := user.Inventory.GetSlot(slot)
	if itemSlot == nil || itemSlot.ObjectID == 0 || itemSlot.Amount < int(amount) {
		return true, nil
	}

	obj := p.ObjectService.GetObject(itemSlot.ObjectID)
	if obj == nil {
		return true, nil
	}

	if itemSlot.Equipped {
		connection.Send(&outgoing.ConsoleMessagePacket{Message: "No puedes vender un objeto equipado.", Font: outgoing.INFO})
		return true, nil
	}

	sellPrice := (obj.Value / 2) * int(amount)
	if sellPrice < 1 { sellPrice = 1 }

	user.Gold += sellPrice
	itemSlot.Amount -= int(amount)
	if itemSlot.Amount <= 0 {
		itemSlot.ObjectID = 0
	}

	connection.Send(&outgoing.ConsoleMessagePacket{
		Message: fmt.Sprintf("Has vendido %d %s por %d monedas de oro.", amount, obj.Name, sellPrice),
		Font:    outgoing.INFO,
	})

	// Update user stats (gold)
	connection.Send(outgoing.NewUpdateUserStatsPacket(user))
	
	// Sync inventory slot
	var updatedObj *model.Object
	if itemSlot.ObjectID != 0 {
		updatedObj = obj
	}
	connection.Send(&outgoing.ChangeInventorySlotPacket{
		Slot:     slotIdx,
		Object:   updatedObj,
		Amount:   itemSlot.Amount,
		Equipped: itemSlot.Equipped,
	})

	return true, nil
}

type CommerceEndPacket struct {
}

func (p *CommerceEndPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	user := connection.GetUser()
	if user != nil {
		user.TradingNPCIndex = 0
	}
	return true, nil
}
