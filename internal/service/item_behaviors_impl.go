package service

import (
	"fmt"
	"time"

	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/utils"
)

// --- Consumables ---

type FoodBehavior struct {
	svc *ItemActionServiceImpl
}

func (b *FoodBehavior) Use(char *model.Character, slot int, obj *model.Object, connection protocol.Connection) {
	// Hunger: 100 is full, 0 is starving.
	char.Hunger = utils.Min(100, char.Hunger+obj.HungerPoints)
	connection.Send(&outgoing.UpdateHungerAndThirstPacket{
		MinHunger: char.Hunger, MaxHunger: 100,
		MinThirst: char.Thirstiness, MaxThirst: 100,
	})
	connection.Send(outgoing.NewUpdateUserStatsPacket(char))
	connection.Send(&outgoing.ConsoleMessagePacket{
		Message: fmt.Sprintf("Has comido %s.", obj.Name),
		Font:    outgoing.INFO,
	})
	b.svc.RemoveOne(char, slot, connection)
}

type DrinkBehavior struct {
	svc *ItemActionServiceImpl
}

func (b *DrinkBehavior) Use(char *model.Character, slot int, obj *model.Object, connection protocol.Connection) {
	// Thirstiness: 100 is full, 0 is thirsty.
	char.Thirstiness = utils.Min(100, char.Thirstiness+obj.ThirstPoints)
	connection.Send(&outgoing.UpdateHungerAndThirstPacket{
		MinHunger: char.Hunger, MaxHunger: 100,
		MinThirst: char.Thirstiness, MaxThirst: 100,
	})
	connection.Send(outgoing.NewUpdateUserStatsPacket(char))
	connection.Send(&outgoing.ConsoleMessagePacket{
		Message: fmt.Sprintf("Has bebido %s.", obj.Name),
		Font:    outgoing.INFO,
	})
	b.svc.RemoveOne(char, slot, connection)
}

// --- Potions ---

type PotionBehavior struct {
	svc *ItemActionServiceImpl
}

func (b *PotionBehavior) Use(char *model.Character, slot int, obj *model.Object, connection protocol.Connection) {
	if char.OriginalAttributes == nil {
		char.OriginalAttributes = make(map[model.Attribute]byte)
		for k, v := range char.Attributes {
			char.OriginalAttributes[k] = v
		}
	}

	switch obj.PotionType {
	case 1: // Agility (Yellow)
		base := char.OriginalAttributes[model.Dexterity]
		modifier := utils.RandomNumber(obj.MinModifier, obj.MaxModifier)
		
		// The new value should always be Base + Modifier, but limited to 40.
		newVal := byte(utils.Min(40, int(base)+modifier))
		
		// ALWAYS POSITIVE: Only set if the new value is higher than current, 
		// OR if the current effect has already expired.
		if newVal > char.Attributes[model.Dexterity] || time.Now().After(char.AgilityEffectEnd) {
			char.Attributes[model.Dexterity] = newVal
		}
		
		char.AgilityEffectEnd = time.Now().Add(time.Duration(obj.Duration) * time.Second)
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "¡Tu agilidad ha aumentado!",
			Font:    outgoing.INFO,
		})
		connection.Send(&outgoing.UpdateStrengthAndDexterityPacket{
			Strength:  char.Attributes[model.Strength],
			Dexterity: char.Attributes[model.Dexterity],
		})

	case 2: // Strength (Green)
		base := char.OriginalAttributes[model.Strength]
		modifier := utils.RandomNumber(obj.MinModifier, obj.MaxModifier)
		
		newVal := byte(utils.Min(40, int(base)+modifier))
		
		if newVal > char.Attributes[model.Strength] || time.Now().After(char.StrengthEffectEnd) {
			char.Attributes[model.Strength] = newVal
		}

		char.StrengthEffectEnd = time.Now().Add(time.Duration(obj.Duration) * time.Second)
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "¡Tu fuerza ha aumentado!",
			Font:    outgoing.INFO,
		})
		connection.Send(&outgoing.UpdateStrengthAndDexterityPacket{
			Strength:  char.Attributes[model.Strength],
			Dexterity: char.Attributes[model.Dexterity],
		})

	case 3: // HP
		modifier := utils.RandomNumber(obj.MinModifier, obj.MaxModifier)
		char.Hp = utils.Min(char.MaxHp, char.Hp+modifier)
	case 4: // Mana
		modifier := utils.RandomNumber(obj.MinModifier, obj.MaxModifier)
		char.Mana = utils.Min(char.MaxMana, char.Mana+modifier)
	case 5: // Poison
		char.Poisoned = false
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "Te has curado del envenenamiento.",
			Font:    outgoing.INFO,
		})
	case 6: // Black Potion (Suicide)
		b.svc.messageService.HandleDeath(char, "¡La poción negra te ha matado!")
		b.svc.RemoveOne(char, slot, connection)
		return
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

// --- Others ---

type MoneyBehavior struct {
	svc *ItemActionServiceImpl
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
	svc *ItemActionServiceImpl
}

func (b *ToolBehavior) Use(char *model.Character, slot int, obj *model.Object, connection protocol.Connection) {
	connection.Send(&outgoing.ConsoleMessagePacket{
		Message: "Has usado la herramienta.",
		Font:    outgoing.INFO,
	})
}

// --- Equipment ---

type EquipGenericBehavior struct {
	svc *ItemActionServiceImpl
	Type model.ObjectType
}

func (b *EquipGenericBehavior) Use(char *model.Character, slot int, obj *model.Object, connection protocol.Connection) {
	b.ToggleEquip(char, slot, obj, connection)
}

func (b *EquipGenericBehavior) ToggleEquip(char *model.Character, slot int, obj *model.Object, connection protocol.Connection) {
	itemSlot := char.Inventory.GetSlot(slot)
	if itemSlot.Equipped {
		itemSlot.Equipped = false
		switch obj.Type {
		case model.OTWeapon:
			char.Weapon = 0
		case model.OTArmor:
			char.Body = b.svc.bodyService.GetBody(char.Race, char.Gender)
		case model.OTShield:
			char.Shield = 0
		case model.OTHelmet:
			char.Helmet = 0
		}
	} else {
		for i := 0; i < model.InventorySlots; i++ {
			s := char.Inventory.GetSlot(i)
			if s.Equipped {
				o := b.svc.objectService.GetObject(s.ObjectID)
				if o != nil && o.Type == obj.Type {
					s.Equipped = false
					b.svc.SyncSlot(char, i, connection)
				}
			}
		}
		itemSlot.Equipped = true
		switch obj.Type {
		case model.OTWeapon:
			char.Weapon = int16(obj.EquippedWeaponGraphic)
		case model.OTArmor:
			if obj.EquippedArmorGraphic > 0 {
				char.Body = obj.EquippedArmorGraphic
			}
		case model.OTShield:
			char.Shield = int16(obj.ID)
		case model.OTHelmet:
			char.Helmet = int16(obj.ID)
		}
	}
	b.svc.SyncSlot(char, slot, connection)
	b.svc.messageService.SendToArea(&outgoing.CharacterChangePacket{Character: char}, char.Position)
}

type BoatBehavior struct {
	svc *ItemActionServiceImpl
}

func (b *BoatBehavior) Use(char *model.Character, slot int, obj *model.Object, connection protocol.Connection) {
	b.ToggleEquip(char, slot, obj, connection)
}

func (b *BoatBehavior) ToggleEquip(char *model.Character, slot int, obj *model.Object, connection protocol.Connection) {
	itemSlot := char.Inventory.GetSlot(slot)
	if char.Sailing {
		char.Sailing = false
		itemSlot.Equipped = false
		connection.Send(&outgoing.NavigateTogglePacket{})
	} else {
		char.Sailing = true
		itemSlot.Equipped = true
		connection.Send(&outgoing.NavigateTogglePacket{})
	}
	b.svc.SyncSlot(char, slot, connection)
	b.svc.messageService.SendToArea(&outgoing.CharacterChangePacket{Character: char}, char.Position)
}
