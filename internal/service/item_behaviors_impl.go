package service

import (
	"fmt"
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/utils"
)

// --- Consumables ---

type FoodBehavior struct {
	svc *ItemActionService
}

func (b *FoodBehavior) Use(char *model.Character, slot int, obj *model.Object, connection protocol.Connection) {
	char.Hunger = utils.Min(100, char.Hunger+obj.HungerPoints)
	connection.Send(outgoing.NewUpdateUserStatsPacket(char))
	connection.Send(&outgoing.ConsoleMessagePacket{
		Message: fmt.Sprintf("Has comido %s.", obj.Name),
		Font:    outgoing.INFO,
	})
	b.svc.RemoveOne(char, slot, connection)
}

type DrinkBehavior struct {
	svc *ItemActionService
}

func (b *DrinkBehavior) Use(char *model.Character, slot int, obj *model.Object, connection protocol.Connection) {
	char.Thirstiness = utils.Min(100, char.Thirstiness+obj.ThirstPoints)
	connection.Send(outgoing.NewUpdateUserStatsPacket(char))
	connection.Send(&outgoing.ConsoleMessagePacket{
		Message: fmt.Sprintf("Has bebido %s.", obj.Name),
		Font:    outgoing.INFO,
	})
	b.svc.RemoveOne(char, slot, connection)
}

type PotionBehavior struct {
	svc *ItemActionService
}

func (b *PotionBehavior) Use(char *model.Character, slot int, obj *model.Object, connection protocol.Connection) {
	switch obj.PotionType {
	case 3: // HP
		modifier := utils.RandomNumber(obj.MinModifier, obj.MaxModifier)
		char.Hp = utils.Min(char.MaxHp, char.Hp+modifier)
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: fmt.Sprintf("Has recuperado %d puntos de vida.", modifier),
			Font:    outgoing.INFO,
		})
	case 4: // Mana
		modifier := utils.RandomNumber(obj.MinModifier, obj.MaxModifier)
		char.Mana = utils.Min(char.MaxMana, char.Mana+modifier)
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
	default:
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "Esta poción no tiene efecto.",
			Font:    outgoing.INFO,
		})
		return
	}
	connection.Send(outgoing.NewUpdateUserStatsPacket(char))
	b.svc.RemoveOne(char, slot, connection)
}

type MoneyBehavior struct {
	svc *ItemActionService
}

func (b *MoneyBehavior) Use(char *model.Character, slot int, obj *model.Object, connection protocol.Connection) {
	itemSlot := char.Inventory.GetSlot(slot)
	goldAmount := itemSlot.Amount
	char.Gold += goldAmount
	connection.Send(&outgoing.ConsoleMessagePacket{
		Message: fmt.Sprintf("Has guardado %d monedas de oro en tu billetera.", goldAmount),
		Font:    outgoing.INFO,
	})
	itemSlot.Amount = 0
	itemSlot.ObjectID = 0
	connection.Send(&outgoing.UpdateGoldPacket{Gold: char.Gold})
	b.svc.SyncSlot(char, slot, connection)
}

type ToolBehavior struct {
	svc *ItemActionService
}

func (b *ToolBehavior) Use(char *model.Character, slot int, obj *model.Object, connection protocol.Connection) {
	// Simple tool identification by type/name for now (mimicking AO)
	var skill model.Skill
	switch obj.ID {
	case 138, 139, 140: // Fishing rods
		skill = model.Fishing
	case 163, 164, 165: // Axes
		skill = model.Lumber
	case 187, 188, 189: // Mining picks
		skill = model.Mining
	case 198, 199, 200: // Carpentry saws
		skill = model.Woodwork
	default:
		// If it's a weapon but not a tool, maybe it's a ranged weapon?
		if obj.Ranged {
			if !char.Inventory.Slots[slot].Equipped {
				connection.Send(&outgoing.ConsoleMessagePacket{
					Message: "Antes de usar el arco deberías equipártelo.",
					Font:    outgoing.INFO,
				})
				return
			}
			connection.Send(&outgoing.SkillRequestTargetPacket{Skill: model.Projectiles})
			return
		}
		
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "No puedes usar este objeto.",
			Font:    outgoing.INFO,
		})
		return
	}

	if !char.Inventory.Slots[slot].Equipped {
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "Debes tener equipada la herramienta para trabajar.",
			Font:    outgoing.INFO,
		})
		return
	}

	connection.Send(&outgoing.SkillRequestTargetPacket{Skill: skill})
}

// --- Equipment ---

type EquipGenericBehavior struct {
	svc     *ItemActionService
	objType model.ObjectType
}

func (b *EquipGenericBehavior) ToggleEquip(char *model.Character, slot int, obj *model.Object, connection protocol.Connection) {
	itemSlot := char.Inventory.GetSlot(slot)
	
	if itemSlot.Equipped {
		itemSlot.Equipped = false
		b.unequip(char, obj)
	} else {
		// Unequip current of same type
		b.unequipCurrent(char, connection)
		itemSlot.Equipped = true
		b.equip(char, obj)
	}

	b.svc.SyncSlot(char, slot, connection)
	
	// Broadcast character change
	b.svc.messageService.SendToArea(&outgoing.CharacterChangePacket{Character: char}, char.Position)
}

func (b *EquipGenericBehavior) equip(char *model.Character, obj *model.Object) {
	switch b.objType {
	case model.OTWeapon:
		char.Weapon = int16(obj.EquippedWeaponGraphic)
	case model.OTArmor:
		char.Body = obj.EquippedArmorGraphic
	case model.OTShield:
		char.Shield = int16(obj.EquippedWeaponGraphic)
	case model.OTHelmet:
		char.Helmet = int16(obj.EquippedHelmetGraphic)
	}
}

func (b *EquipGenericBehavior) unequip(char *model.Character, obj *model.Object) {
	switch b.objType {
	case model.OTWeapon:
		char.Weapon = 2
	case model.OTArmor:
		char.Body = b.svc.bodyService.GetBody(char.Race, char.Gender)
	case model.OTShield:
		char.Shield = 2
	case model.OTHelmet:
		char.Helmet = 2
	}
}

func (b *EquipGenericBehavior) unequipCurrent(char *model.Character, conn protocol.Connection) {
	for i := 0; i < model.InventorySlots; i++ {
		s := char.Inventory.GetSlot(i)
		if s.Equipped {
			o := b.svc.objectService.GetObject(s.ObjectID)
			if o != nil && o.Type == b.objType {
				s.Equipped = false
				b.svc.SyncSlot(char, i, conn)
			}
		}
	}
}
