package service

import (
	"fmt"
	"strings"

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

// --- Boat ---

type BoatBehavior struct {
	svc *ItemActionService
}

func (b *BoatBehavior) ToggleEquip(char *model.Character, slot int, obj *model.Object, connection protocol.Connection) {
	// 1. Mount Check
	// Note: Assuming 'Equitando' equivalent is a flag or checks against 'OTMount' (not implemented in model yet, but sticking to navigating.txt logic)
	// If future mount system is added, add check here: if char.Mounted { ... }

	// 2. Class/Level Checks (UseInvItem logic)
	// "Verifica si esta aproximado al agua antes de permitirle navegar" logic combined with Level checks
	// Requirements for sailing (only checked when STARTING to sail)
	if !char.Sailing {
		minLevel := 25
		if char.Archetype == model.Worker || char.Archetype == model.Pirate {
			minLevel = 20
		}

		if int(char.Level) < minLevel {
			connection.Send(&outgoing.ConsoleMessagePacket{
				Message: fmt.Sprintf("Para recorrer los mares debes ser nivel %d o superior.", minLevel),
				Font:    outgoing.INFO,
			})
			return
		}

		// Worker 20-25 needs 100 Fishing
		if char.Archetype == model.Worker && int(char.Level) >= 20 && int(char.Level) < 25 {
			// Assuming Fishing skill exists and is checked.
			if char.Skills[model.Fishing] < 100 {
				connection.Send(&outgoing.ConsoleMessagePacket{
					Message: "Para recorrer los mares siendo trabajador nivel 20-25 necesitas 100 de pesca.",
					Font:    outgoing.INFO,
				})
				return
			}
		}
	}

	// 3. Proximity Checks & Toggle Logic
	isNearWater := b.isNearWater(char)
	isNearLand := b.isNearLand(char)

	if !char.Sailing {
		// Trying to SAIL
		if !isNearWater {
			connection.Send(&outgoing.ConsoleMessagePacket{
				Message: "Debes aproximarte al agua para usar un barco.",
				Font:    outgoing.INFO,
			})
			return
		}

		// Check Boat Restrictions (DoNavega logic)
		if !b.canUseBoatType(char, obj, connection) {
			return
		}

		// Check Navigation Skill
		// Worker: 60, Others: obj.MinSkill (not in Object model yet, assuming standard or 0/60)
		reqSkill := 0 // Default
		// If Object had MinSkill field we would use it. For now, let's assume standard behavior or add a placeholder.
		// navegar.txt: "SkillNecesario = IIf(.Clase = eClass.Worker, 60, Barco.MinSkill)"
		// Since we don't have Barco.MinSkill in model.Object, we'll assume 60 for everyone or 0 if not defined.
		// Let's implement a safe default or checking 'Navegacion' skill.
		// Usually navigation needs some points. Let's say 60 for everyone for high tier boats?
		// For basic "Barca", maybe 0?
		// "EsGalera" -> likely higher skill.
		// I will just implement the Worker exception as requested and skip MinSkill for others if not available.
		
		skillVal := char.Skills[model.Sailing] // Navegacion
		reqSkill = 0 // Basic boat
		if char.Archetype == model.Worker {
			reqSkill = 60
		}
		// If we wanted to be strict:
		// if obj.Name == "Galera" { reqSkill = ... }

		if skillVal < reqSkill {
			connection.Send(&outgoing.ConsoleMessagePacket{
				Message: "No tienes suficientes conocimientos para usar este barco.",
				Font:    outgoing.INFO,
			})
			return
		}

		// Equip / Start Sailing
		char.Sailing = true
		
		// Visuals
		char.OriginalHead = char.Head // Backup head (though logic might be simpler: head=0)
		char.Head = 0
		
		// Body Change
		if char.Dead {
			// iFragataFantasmal - usually hardcoded or specific ID. 
			// Let's assume a placeholder or standard ID. In standard AO 0.13: 86? 
			// I'll leave current body if dead or use a known ghost ship ID if found.
			// For now, keep current or set to obj.EquippedArmorGraphic (which presumably is the boat)
			char.Body = obj.EquippedArmorGraphic
		} else {
			char.Body = obj.EquippedArmorGraphic
		}

		// Update Inventory (Mark as equipped)
		itemSlot := char.Inventory.GetSlot(slot)
		itemSlot.Equipped = true
		b.svc.SyncSlot(char, slot, connection)

		// Send Navigate Toggle
		connection.Send(&outgoing.NavigateTogglePacket{})

		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "¡Has comenzado a navegar!",
			Font:    outgoing.INFO,
		})

	} else {
		// Trying to DOCK
		// Check if it's the SAME boat we are using? 
		// Usually slot.Equipped is true.

		if !isNearLand {
			connection.Send(&outgoing.ConsoleMessagePacket{
				Message: "Debes aproximarte a la tierra para desembarcar.",
				Font:    outgoing.INFO,
			})
			return
		}

		// Unequip / Stop Sailing
		char.Sailing = false
		
		// Restore Visuals
		// Need to restore original body and head.
		// Head:
		// Logic says: If !Invisible -> Head = OrigHead.
		if !char.Invisible {
			// We need to know what the original head was.
			// `char.OriginalHead` might be useful if we tracked it, 
			// but usually we can just rely on `char.Head` being restored if we saved it.
			// But wait, `char.Head` was 0.
			// In `model.Character`, `OriginalHead` exists. Assuming it holds the base head.
			char.Head = char.OriginalHead 
		}

		// Body:
		// Restore body based on Race/Gender (naked) or Armor?
		// If Armor is equipped, we should show armor.
		// Problem: We don't know if armor was equipped.
		// Wait, in AO you unequip armor to equip boat? No, boat overrides body.
		// When unequipped, we should check if Armor is equipped in inventory.
		
		restoredBody := b.svc.bodyService.GetBody(char.Race, char.Gender) // Default naked
		
		// Find equipped armor
		for i := 0; i < model.InventorySlots; i++ {
			s := char.Inventory.GetSlot(i)
			if s.Equipped {
				o := b.svc.objectService.GetObject(s.ObjectID)
				if o != nil && o.Type == model.OTArmor {
					restoredBody = o.EquippedArmorGraphic
					break
				}
			}
		}
		char.Body = restoredBody

		// Update Inventory
		itemSlot := char.Inventory.GetSlot(slot)
		itemSlot.Equipped = false
		b.svc.SyncSlot(char, slot, connection)
		
		// Send Navigate Toggle
		connection.Send(&outgoing.NavigateTogglePacket{})
		
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "Has dejado de navegar.",
			Font:    outgoing.INFO,
		})
	}

	// Update Character on clients
	b.svc.messageService.SendToArea(&outgoing.CharacterChangePacket{Character: char}, char.Position)
}

func (b *BoatBehavior) isNearWater(char *model.Character) bool {
	return b.checkSurroundings(char, func(t *model.Tile) bool {
		return t.IsWater
	})
}

func (b *BoatBehavior) isNearLand(char *model.Character) bool {
	return b.checkSurroundings(char, func(t *model.Tile) bool {
		hasBridge := t.Layer2 > 0 || t.Layer3 > 0
		return !t.IsWater || hasBridge
	})
}

func (b *BoatBehavior) checkSurroundings(char *model.Character, predicate func(*model.Tile) bool) bool {
	m := b.svc.messageService.MapService.GetMap(char.Position.Map)
	if m == nil {
		return false
	}
	x, y := int(char.Position.X), int(char.Position.Y)

	// Check 4 directions: N, S, E, W
	coords := [][2]int{{x, y - 1}, {x, y + 1}, {x + 1, y}, {x - 1, y}}
	
	for _, c := range coords {
		nx, ny := c[0], c[1]
		if nx >= 0 && nx < model.MapWidth && ny >= 0 && ny < model.MapHeight {
			tile := m.GetTile(nx, ny)
			if predicate(tile) {
				return true
			}
		}
	}
	return false
}

func (b *BoatBehavior) canUseBoatType(char *model.Character, obj *model.Object, conn protocol.Connection) bool {
	name := strings.ToUpper(obj.Name)

	if strings.Contains(name, "GALERA") {
		allowed := char.Archetype == model.Assasin ||
			char.Archetype == model.Pirate ||
			char.Archetype == model.Bandit ||
			char.Archetype == model.Cleric ||
			char.Archetype == model.Thief ||
			char.Archetype == model.Paladin
		
		if !allowed {
			conn.Send(&outgoing.ConsoleMessagePacket{
				Message: "Solo los Piratas, Asesinos, Bandidos, Clérigos, Ladrones y Paladines pueden usar Galera!!",
				Font:    outgoing.INFO,
			})
			return false
		}
	}

	if strings.Contains(name, "GALEON") {
		allowed := char.Archetype == model.Thief || char.Archetype == model.Pirate
		
		if !allowed {
			conn.Send(&outgoing.ConsoleMessagePacket{
				Message: "Solo los Ladrones y Piratas pueden usar Galeón!!",
				Font:    outgoing.INFO,
			})
			return false
		}
	}

	return true
}
