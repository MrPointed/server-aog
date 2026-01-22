package service

import (
	"fmt"
	"math/rand"

	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type CombatService struct {
	messageService *MessageService
	objectService  *ObjectService
	mapService     *MapService
}

func NewCombatService(messageService *MessageService, objectService *ObjectService, mapService *MapService) *CombatService {
	return &CombatService{
		messageService: messageService,
		objectService:  objectService,
		mapService:     mapService,
	}
}

func (s *CombatService) ResolveAttack(attacker *model.Character, target any) {
	switch t := target.(type) {
	case *model.Character:
		s.resolvePVP(attacker, t)
	case *model.WorldNPC:
		s.resolvePVE(attacker, t)
	}
}

func (s *CombatService) resolvePVP(attacker *model.Character, target *model.Character) {
	if target.Dead {
		return
	}

	if !s.checkHit(attacker, target) {
		s.messageService.userService.GetConnection(attacker).Send(&outgoing.ConsoleMessagePacket{
			Message: "¡Has fallado el golpe!",
			Font:    outgoing.INFO,
		})
		s.messageService.userService.GetConnection(target).Send(&outgoing.ConsoleMessagePacket{
			Message: fmt.Sprintf("¡%s ha fallado el golpe!", attacker.Name),
			Font:    outgoing.INFO,
		})
		return
	}

	damage := s.calculateDamage(attacker)
	target.Hp -= damage
	if target.Hp < 0 {
		target.Hp = 0
	}

	connTarget := s.messageService.userService.GetConnection(target)
	if connTarget != nil {
		connTarget.Send(outgoing.NewUpdateUserStatsPacket(target))
		connTarget.Send(&outgoing.ConsoleMessagePacket{
			Message: fmt.Sprintf("¡%s te ha golpeado por %d puntos de daño!", attacker.Name, damage),
			Font:    outgoing.INFO,
		})
	}

	connAttacker := s.messageService.userService.GetConnection(attacker)
	if connAttacker != nil {
		connAttacker.Send(&outgoing.ConsoleMessagePacket{
			Message: fmt.Sprintf("¡Has golpeado a %s por %d puntos de daño!", target.Name, damage),
			Font:    outgoing.INFO,
		})
	}

	if target.Hp <= 0 {
		s.handleCharacterDeath(target)
	}
}

func (s *CombatService) handleCharacterDeath(char *model.Character) {
	char.Dead = true
	char.Hp = 0
	char.Body = 8   // Casper
	char.Head = 500 // Casper head

	conn := s.messageService.userService.GetConnection(char)
	if conn != nil {
		conn.Send(outgoing.NewUpdateUserStatsPacket(char))
		conn.Send(&outgoing.ConsoleMessagePacket{
			Message: "¡Has muerto!",
			Font:    outgoing.INFO,
		})
	}

	// Broadcast change
	s.messageService.SendToArea(&outgoing.CharacterChangePacket{Character: char}, char.Position)
}

func (s *CombatService) resolvePVE(attacker *model.Character, target *model.WorldNPC) {
	// Simple hit chance for NPC: Fixed evasion for now
	// TODO: Use NPC stats when available
	if rand.Float32() > 0.9 { // 90% hit chance on NPCs for now
		s.messageService.userService.GetConnection(attacker).Send(&outgoing.ConsoleMessagePacket{
			Message: "¡Has fallado el golpe!",
			Font:    outgoing.INFO,
		})
		return
	}

	// Calculate damage
	damage := s.calculateDamage(attacker)
	
	// Apply to NPC
	target.HP -= damage
	
	s.messageService.userService.GetConnection(attacker).Send(&outgoing.ConsoleMessagePacket{
		Message: fmt.Sprintf("¡Has golpeado a la criatura por %d puntos de daño!", damage),
		Font:    outgoing.INFO,
	})

	if target.HP <= 0 {
		s.handleNpcDeath(attacker, target)
	}
}

func (s *CombatService) checkHit(attacker *model.Character, target *model.Character) bool {
	// Attacker Power
	weaponSkill := 0
	equippedWeapon := s.getEquippedWeapon(attacker)
	if equippedWeapon != nil {
		if equippedWeapon.Ranged {
			weaponSkill = attacker.Skills[model.Projectiles]
		} else {
			weaponSkill = attacker.Skills[model.CombatTactics]
		}
	} else {
		weaponSkill = attacker.Skills[model.Wrestling]
	}

	dex := int(attacker.Attributes[model.Dexterity])
	attackPower := float64(weaponSkill) + float64(dex)*0.7

	// Defender Power
	defenseSkill := 0
	equippedShield := s.getEquippedShield(target)
	if equippedShield != nil {
		defenseSkill = target.Skills[model.Defense] // Escudos? usually Defense involves shield
	} else {
		defenseSkill = target.Skills[model.CombatTactics] // Parrying with weapon? Or Wrestling?
		// Simplified: Use CombatTactics or Wrestling as partial defense
		if s.getEquippedWeapon(target) == nil {
			defenseSkill = target.Skills[model.Wrestling]
		} else {
			defenseSkill = target.Skills[model.CombatTactics]
		}
	}
	
	targetDex := int(target.Attributes[model.Dexterity])
	defensePower := float64(defenseSkill) + float64(targetDex)*0.7

	// Chance
	// Simple ratio: Power / (Power + Defense)
	// Or standard AO: (Hit - Eva) > Random
	
	// Let's use a standard 0-100 check
	// Maximum skill is 100. Max Dex ~30-40.
	// Max Power ~130.
	
	// Diff factor
	diff := int(attackPower - defensePower)
	
	// Base chance 50%
	chance := 50 + diff
	
	if chance < 10 {
		chance = 10
	}
	if chance > 90 {
		chance = 90
	}

	return rand.Intn(100) < chance
}

func (s *CombatService) getEquippedWeapon(char *model.Character) *model.Object {
	for i := 0; i < model.InventorySlots; i++ {
		slot := char.Inventory.GetSlot(i)
		if slot.Equipped {
			obj := s.objectService.GetObject(slot.ObjectID)
			if obj != nil && obj.Type == model.OTWeapon {
				return obj
			}
		}
	}
	return nil
}

func (s *CombatService) getEquippedShield(char *model.Character) *model.Object {
	for i := 0; i < model.InventorySlots; i++ {
		slot := char.Inventory.GetSlot(i)
		if slot.Equipped {
			obj := s.objectService.GetObject(slot.ObjectID)
			if obj != nil && obj.Type == model.OTShield {
				return obj
			}
		}
	}
	return nil
}

func (s *CombatService) calculateDamage(attacker *model.Character) int {
	// Base damage from attributes
	strength := int(attacker.Attributes[model.Strength])
	
	minDmg := strength / 3
	maxDmg := strength / 2

	// Weapon bonus
	obj := s.getEquippedWeapon(attacker)
	if obj != nil {
		minDmg += obj.MinHit
		maxDmg += obj.MaxHit
	} else {
		// Hand to hand
		minDmg += 1
		maxDmg += 2
	}

	baseDamage := minDmg + rand.Intn(maxDmg-minDmg+1)
	
	// Apply Class Modifier
	mod := s.getClassDamageModifier(attacker.Archetype)
	
	return int(float32(baseDamage) * mod)
}

func (s *CombatService) getClassDamageModifier(class model.UserArchetype) float32 {
	switch class {
	case model.Mage:
		return 0.4
	case model.Cleric, model.Bard, model.Druid:
		return 0.6
	case model.Assasin:
		return 0.9 // Backstab logic needed later
	case model.Thief:
		return 0.6
	case model.Warrior:
		return 1.0
	case model.Paladin:
		return 0.9
	case model.Hunter:
		return 0.85
	case model.Worker:
		return 0.4
	case model.Pirate, model.Bandit:
		return 0.9
	default:
		return 1.0
	}
}

func (s *CombatService) handleNpcDeath(killer *model.Character, npc *model.WorldNPC) {
	s.messageService.SendToArea(&outgoing.CharacterRemovePacket{CharIndex: npc.Index}, npc.Position)
	
	killer.Exp += npc.NPC.Exp
	s.messageService.userService.GetConnection(killer).Send(&outgoing.ConsoleMessagePacket{
		Message: fmt.Sprintf("¡Has matado a la criatura! Ganaste %d puntos de experiencia.", npc.NPC.Exp),
		Font:    outgoing.INFO,
	})
	
	// Sync stats
	s.messageService.userService.GetConnection(killer).Send(outgoing.NewUpdateUserStatsPacket(killer))
	
	// Drop items
	for _, drop := range npc.NPC.Drops {
		obj := s.objectService.GetObject(drop.ObjectID)
		if obj != nil {
			worldObj := &model.WorldObject{
				Object: obj,
				Amount: drop.Amount,
			}
			// Only drop if tile is empty
			if s.mapService.GetObjectAt(npc.Position) == nil {
				s.mapService.PutObject(npc.Position, worldObj)
				s.messageService.SendToArea(&outgoing.ObjectCreatePacket{
					X:            npc.Position.X,
					Y:            npc.Position.Y,
					GraphicIndex: int16(obj.GraphicIndex),
				}, npc.Position)
			}
		}
	}

	// Remove from map
	s.mapService.RemoveNPC(npc)
}
