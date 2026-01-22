package incoming

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/service"
	"github.com/ao-go-server/internal/model"
)

type EquipItemPacket struct {
	ObjectService  *service.ObjectService
	MessageService *service.MessageService
	BodyService    *service.CharacterBodyService
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

	if char.Dead {
		connection.Send(&outgoing.ConsoleMessagePacket{Message: "¡Estás muerto!", Font: outgoing.INFO})
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

	// Toggle equipment
	if itemSlot.Equipped {
		itemSlot.Equipped = false
		p.unequip(char, obj)
	} else {
		// Unequip current of same type
		p.unequipCurrent(char, obj.Type, connection)
		itemSlot.Equipped = true
		p.equip(char, obj)
	}

	// Sync inventory slot
	connection.Send(&outgoing.ChangeInventorySlotPacket{
		Slot:     slotIdx,
		Object:   obj,
		Amount:   itemSlot.Amount,
		Equipped: itemSlot.Equipped,
	})

	// Broadcast character change
	p.MessageService.SendToArea(&outgoing.CharacterChangePacket{Character: char}, char.Position)

	return true, nil
}

func (p *EquipItemPacket) unequip(char *model.Character, obj *model.Object) {
	switch obj.Type {
	case model.OTWeapon: char.Weapon = 2 // Default/None
	case model.OTArmor: 
		char.Body = p.BodyService.GetBody(char.Race, char.Gender)
	case model.OTShield: char.Shield = 2
	case model.OTHelmet: char.Helmet = 2
	}
}

func (p *EquipItemPacket) equip(char *model.Character, obj *model.Object) {
	switch obj.Type {
	case model.OTWeapon: char.Weapon = int16(obj.EquippedWeaponGraphic)
	case model.OTArmor: char.Body = obj.EquippedBodyGraphic
	case model.OTShield: char.Shield = int16(obj.EquippedWeaponGraphic) // AO uses same field for shield grh
	case model.OTHelmet: char.Helmet = int16(obj.EquippedHeadGraphic)
	}
}

func (p *EquipItemPacket) unequipCurrent(char *model.Character, objType model.ObjectType, conn protocol.Connection) {
	for i := 0; i < model.InventorySlots; i++ {
		s := char.Inventory.GetSlot(i)
		if s.Equipped {
			o := p.ObjectService.GetObject(s.ObjectID)
			if o != nil && o.Type == objType {
				s.Equipped = false
				// Sync this slot too
				conn.Send(&outgoing.ChangeInventorySlotPacket{
					Slot:     byte(i + 1),
					Object:   o,
					Amount:   s.Amount,
					Equipped: false,
				})
			}
		}
	}
}
