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

	// 80% hit chance for now
	if rand.Float32() > 0.8 {
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
	// Simple hit chance: (Skill + Luck) vs (NPC evasion)
	// For now, let's just make it 80% chance
	if rand.Float32() > 0.8 {
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

func (s *CombatService) calculateDamage(attacker *model.Character) int {
	// Base damage from attributes
	strength := int(attacker.Attributes[model.Strength])
	
	minDmg := strength / 3
	maxDmg := strength / 2

	// Weapon bonus
	weaponID := 0
	for i := 0; i < model.InventorySlots; i++ {
		slot := attacker.Inventory.GetSlot(i)
		if slot.Equipped {
			obj := s.objectService.GetObject(slot.ObjectID)
			if obj != nil && obj.Type == model.OTWeapon {
				minDmg += obj.MinHit
				maxDmg += obj.MaxHit
				weaponID = obj.ID
				break
			}
		}
	}

	if weaponID == 0 {
		// Hand to hand
		minDmg += 1
		maxDmg += 2
	}

	damage := minDmg + rand.Intn(maxDmg-minDmg+1)
	return damage
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
