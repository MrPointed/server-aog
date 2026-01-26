package service

import (
	"fmt"

	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type ItemActionServiceImpl struct {
	objectService   ObjectService
	messageService  MessageService
	intervalService IntervalService
	bodyService     BodyService

	useBehaviors   map[model.ObjectType]ItemBehavior
	equipBehaviors map[model.ObjectType]EquipBehavior
}

func NewItemActionServiceImpl(objSvc ObjectService, msgSvc MessageService, intSvc IntervalService, bodySvc BodyService) ItemActionService {
	s := &ItemActionServiceImpl{
		objectService:   objSvc,
		messageService:  msgSvc,
		intervalService: intSvc,
		bodyService:     bodySvc,
		useBehaviors:    make(map[model.ObjectType]ItemBehavior),
		equipBehaviors:  make(map[model.ObjectType]EquipBehavior),
	}
	s.registerDefaultBehaviors()
	return s
}

func (s *ItemActionServiceImpl) UseItem(char *model.Character, slot int, connection protocol.Connection) {
	if char == nil || char.Dead {
		return
	}

	itemSlot := char.Inventory.GetSlot(slot)
	if itemSlot == nil || itemSlot.ObjectID == 0 {
		return
	}

	obj := s.objectService.GetObject(itemSlot.ObjectID)
	if obj == nil {
		return
	}

	if !s.intervalService.CanUseItem(char) {
		return
	}

	behavior, ok := s.useBehaviors[obj.Type]
	if !ok {
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "No puedes usar este objeto.",
			Font:    outgoing.INFO,
		})
		return
	}

	// Requirements check
	if !s.CanUse(char, obj, connection) {
		return
	}

	behavior.Use(char, slot, obj, connection)
	s.intervalService.UpdateLastItem(char)
}

func (s *ItemActionServiceImpl) EquipItem(char *model.Character, slot int, connection protocol.Connection) {
	if char == nil || char.Dead {
		return
	}

	itemSlot := char.Inventory.GetSlot(slot)
	if itemSlot == nil || itemSlot.ObjectID == 0 {
		return
	}

	obj := s.objectService.GetObject(itemSlot.ObjectID)
	if obj == nil {
		return
	}

	behavior, ok := s.equipBehaviors[obj.Type]
	if !ok {
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "No puedes equipar este objeto.",
			Font:    outgoing.INFO,
		})
		return
	}

	// Requirements check
	if !s.CanEquip(char, obj, connection) {
		return
	}

	behavior.ToggleEquip(char, slot, obj, connection)
}

func (s *ItemActionServiceImpl) CanUse(char *model.Character, obj *model.Object, connection protocol.Connection) bool {

	// Newbie check
	if obj.Newbie && char.Level > 12 {
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "Solo los newbies pueden usar este objeto.",
			Font:    outgoing.INFO,
		})
		return false
	}

	// Level check
	if int(char.Level) < obj.MinLevel {
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: fmt.Sprintf("Necesitas ser nivel %d para poder usar este objeto.", obj.MinLevel),
			Font:    outgoing.INFO,
		})
		return false
	}

	// Skill check (if implemented in model)
	// TODO: Add skill check when SkillRequerido is added to model.Object

	return true
}

func (s *ItemActionServiceImpl) CanEquip(char *model.Character, obj *model.Object, connection protocol.Connection) bool {
	// Dead check
	if char.Dead {
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "¡Estás muerto!",
			Font:    outgoing.INFO,
		})
		return false
	}

	if s.messageService.MapService().IsInvalidPosition(char.Position) {
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "Posición inválida.",
			Font:    outgoing.INFO,
		})
		return false
	}

	// Newbie check
	if obj.Newbie && char.Level > 12 {
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "Solo los newbies pueden usar este objeto.",
			Font:    outgoing.INFO,
		})
		return false
	}

	// Level check
	if int(char.Level) < obj.MinLevel {
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: fmt.Sprintf("Necesitas ser nivel %d para poder equipar este objeto.", obj.MinLevel),
			Font:    outgoing.INFO,
		})
		return false
	}

	// Gender check
	if obj.OnlyMen && char.Gender != model.Male {
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "Tu género no puede usar este objeto.",
			Font:    outgoing.INFO,
		})
		return false
	}
	if obj.OnlyWomen && char.Gender != model.Female {
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "Tu género no puede usar este objeto.",
			Font:    outgoing.INFO,
		})
		return false
	}

	// Race check
	isDwarfOrGnome := char.Race == model.Dwarf || char.Race == model.Gnome
	if obj.DwarfOnly && !isDwarfOrGnome {
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "Tu raza no puede usar este objeto.",
			Font:    outgoing.INFO,
		})
		return false
	}
	if obj.DarkElfOnly && char.Race != model.DarkElf {
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "Tu raza no puede usar este objeto.",
			Font:    outgoing.INFO,
		})
		return false
	}
	if !obj.DwarfOnly && !obj.DarkElfOnly && isDwarfOrGnome && obj.Type == model.OTArmor {
		// Only apply "non-dwarf armor" restriction to armor (typical AO logic)
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "Tu raza no puede usar este objeto.",
			Font:    outgoing.INFO,
		})
		return false
	}

	// Class check
	for _, forbidden := range obj.ForbiddenArchetypes {
		if forbidden == char.Archetype {
			connection.Send(&outgoing.ConsoleMessagePacket{
				Message: "Tu clase no puede usar este objeto.",
				Font:    outgoing.INFO,
			})
			return false
		}
	}

	// TODO: Add Gender, Race, alignment checks from inventario_vb

	return true
}

func (s *ItemActionServiceImpl) registerDefaultBehaviors() {
	// Use behaviors
	s.useBehaviors[model.OTFood] = &FoodBehavior{s}
	s.useBehaviors[model.OTDrink] = &DrinkBehavior{s}
	s.useBehaviors[model.OTPotion] = &PotionBehavior{s}
	s.useBehaviors[model.OTMoney] = &MoneyBehavior{s}
	s.useBehaviors[model.OTWeapon] = &ToolBehavior{s}

	// Equip behaviors
	weaponBehavior := &EquipGenericBehavior{s, model.OTWeapon}
	armorBehavior := &EquipGenericBehavior{s, model.OTArmor}
	shieldBehavior := &EquipGenericBehavior{s, model.OTShield}
	helmetBehavior := &EquipGenericBehavior{s, model.OTHelmet}
	ringBehavior := &EquipGenericBehavior{s, model.OTRing}

	s.equipBehaviors[model.OTWeapon] = weaponBehavior
	s.equipBehaviors[model.OTArmor] = armorBehavior
	s.equipBehaviors[model.OTShield] = shieldBehavior
	s.equipBehaviors[model.OTHelmet] = helmetBehavior
	s.equipBehaviors[model.OTRing] = ringBehavior
	s.equipBehaviors[model.OTBoat] = &BoatBehavior{s}
}

// Internal sync helper
func (s *ItemActionServiceImpl) SyncSlot(char *model.Character, slot int, connection protocol.Connection) {
	itemSlot := char.Inventory.GetSlot(slot)
	var obj *model.Object
	if itemSlot.ObjectID != 0 {
		obj = s.objectService.GetObject(itemSlot.ObjectID)
	}

	connection.Send(&outgoing.ChangeInventorySlotPacket{
		Slot:     byte(slot + 1),
		Object:   obj,
		Amount:   itemSlot.Amount,
		Equipped: itemSlot.Equipped,
	})
}

func (s *ItemActionServiceImpl) RemoveOne(char *model.Character, slot int, connection protocol.Connection) {
	itemSlot := char.Inventory.GetSlot(slot)
	itemSlot.Amount--
	if itemSlot.Amount <= 0 {
		if itemSlot.Equipped {
			// This shouldn't happen usually for consumables but good to have
			itemSlot.Equipped = false
		}
		itemSlot.ObjectID = 0
	}
	s.SyncSlot(char, slot, connection)
}
