package service

import (
	"fmt"
	"math/rand"

	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/utils"
)

type CombatService struct {
	messageService *MessageService
	objectService  *ObjectService
	mapService     *MapService
	formulas       *CombatFormulas
	intervals      *IntervalService
}

func NewCombatService(messageService *MessageService, objectService *ObjectService, mapService *MapService, formulas *CombatFormulas, intervals *IntervalService) *CombatService {
	return &CombatService{
		messageService: messageService,
		objectService:  objectService,
		mapService:     mapService,
		formulas:       formulas,
		intervals:      intervals,
	}
}

func (s *CombatService) ResolveAttack(attacker *model.Character, target any) {
	if attacker.Dead {
		s.messageService.SendConsoleMessage(attacker, "¡Estás muerto!", outgoing.INFO)
		return
	}

	// Check stamina
	if attacker.Stamina < 10 {
		s.messageService.SendConsoleMessage(attacker, "Estás muy cansado para luchar.", outgoing.INFO)
		return
	}

	// Check interval
	if !s.intervals.CanAttack(attacker) {
		return
	}

	switch t := target.(type) {
	case *model.Character:
		s.resolvePVP(attacker, t)
	case *model.WorldNPC:
		s.resolvePVE(attacker, t)
	}
	
	// Update last attack time
	s.intervals.UpdateLastAttack(attacker)

	// Consume stamina
	attacker.Stamina -= utils.RandomNumber(1, 10)
	if attacker.Stamina < 0 { attacker.Stamina = 0 }
	s.messageService.userService.GetConnection(attacker).Send(outgoing.NewUpdateUserStatsPacket(attacker))
}

func (s *CombatService) resolvePVP(attacker *model.Character, victim *model.Character) {
	if victim.Dead {
		s.messageService.SendConsoleMessage(attacker, "No puedes atacar a un espíritu.", outgoing.INFO)
		return
	}

	weapon := s.getEquippedWeapon(attacker)
	
	// Hit check
	attackerPower := s.formulas.GetAttackPower(attacker, weapon)
	victimEvasion := s.formulas.GetEvasionPower(victim)
	
	// Shield bonus
	if s.getEquippedShield(victim) != nil {
		victimEvasion += s.formulas.GetShieldEvasionPower(victim)
	}

	chance := s.formulas.CalculateHitChance(attackerPower, victimEvasion)
	
	if rand.Intn(100) >= chance {
		s.messageService.SendConsoleMessage(attacker, "¡Has fallado el golpe!", outgoing.FIGHT)
		s.messageService.SendConsoleMessage(victim, fmt.Sprintf("¡%s ha fallado el golpe!", attacker.Name), outgoing.FIGHT)
		return
	}

	// Damage calculation
	damage := s.formulas.CalculateDamage(attacker, weapon, false)
	
	// Armor defense
	armor := s.getEquippedArmor(victim)
	if armor != nil {
		defense := utils.RandomNumber(armor.MinDef, armor.MaxDef)
		damage -= defense
	}
	
	if damage < 1 { damage = 1 }

	victim.Hp -= damage
	if victim.Hp < 0 { victim.Hp = 0 }

	// Feedback
	s.messageService.SendConsoleMessage(attacker, fmt.Sprintf("¡Has golpeado a %s por %d!", victim.Name, damage), outgoing.FIGHT)
	s.messageService.SendConsoleMessage(victim, fmt.Sprintf("¡%s te ha golpeado por %d!", attacker.Name, damage), outgoing.FIGHT)
	
	// Blood FX
	s.messageService.SendToArea(&outgoing.CreateFxPacket{
		CharIndex: victim.CharIndex,
		FxID:      1, // Blood placeholder
		Loops:     0,
	}, victim.Position)

	if victim.Hp <= 0 {
		s.handleCharacterDeath(victim)
	} else {
		connVictim := s.messageService.userService.GetConnection(victim)
		if connVictim != nil {
			connVictim.Send(outgoing.NewUpdateUserStatsPacket(victim))
		}
	}
}

func (s *CombatService) resolvePVE(attacker *model.Character, victim *model.WorldNPC) {
	weapon := s.getEquippedWeapon(attacker)
	
	// Hit check
	attackerPower := s.formulas.GetAttackPower(attacker, weapon)
	victimEvasion := victim.NPC.EvasionPower
	
	chance := s.formulas.CalculateHitChance(attackerPower, victimEvasion)
	
	if rand.Intn(100) >= chance {
		s.messageService.SendConsoleMessage(attacker, "¡Has fallado el golpe!", outgoing.FIGHT)
		return
	}

	// Damage
	damage := s.formulas.CalculateDamage(attacker, weapon, true)
	
	// NPC Defense
	damage -= victim.NPC.Defense
	
	if damage < 1 { damage = 1 }

	victim.HP -= damage
	s.messageService.SendConsoleMessage(attacker, fmt.Sprintf("¡Has golpeado a la criatura por %d!", damage), outgoing.FIGHT)

	if victim.HP <= 0 {
		s.handleNpcDeath(attacker, victim)
	}
}

func (s *CombatService) handleCharacterDeath(char *model.Character) {
	char.Dead = true
	char.Hp = 0
	char.Body = 8   // Ghost
	char.Head = 500 // Ghost head

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

func (s *CombatService) handleNpcDeath(killer *model.Character, npc *model.WorldNPC) {
	s.messageService.SendToArea(&outgoing.CharacterRemovePacket{CharIndex: npc.Index}, npc.Position)
	
	killer.Exp += npc.NPC.Exp
	s.messageService.SendConsoleMessage(killer, fmt.Sprintf("¡Has matado a la criatura! Ganaste %d exp.", npc.NPC.Exp), outgoing.INFO)
	
	s.messageService.userService.GetConnection(killer).Send(outgoing.NewUpdateUserStatsPacket(killer))
	
	// Drop logic
	for _, drop := range npc.NPC.Drops {
		obj := s.objectService.GetObject(drop.ObjectID)
		if obj != nil {
			worldObj := &model.WorldObject{
				Object: obj,
				Amount: drop.Amount,
			}
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

	s.mapService.RemoveNPC(npc)
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

func (s *CombatService) getEquippedArmor(char *model.Character) *model.Object {
	for i := 0; i < model.InventorySlots; i++ {
		slot := char.Inventory.GetSlot(i)
		if slot.Equipped {
			obj := s.objectService.GetObject(slot.ObjectID)
			if obj != nil && obj.Type == model.OTArmor {
				return obj
			}
		}
	}
	return nil
}