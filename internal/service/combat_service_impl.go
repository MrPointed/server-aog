package service

import (
	"fmt"
	"math/rand"

	"github.com/ao-go-server/internal/config"
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/utils"
)

type CombatServiceImpl struct {
	messageService  MessageService
	objectService   ObjectService
	npcService      NpcService
	mapService      MapService
	formulas        *CombatFormulas
	intervals       IntervalService
	trainingService TrainingService
	config          *config.Config
}

func NewCombatServiceImpl(messageService MessageService, objectService ObjectService, npcService NpcService, mapService MapService, formulas *CombatFormulas, intervals IntervalService, trainingService TrainingService, cfg *config.Config) CombatService {
	return &CombatServiceImpl{
		messageService:  messageService,
		objectService:   objectService,
		npcService:      npcService,
		mapService:      mapService,
		formulas:        formulas,
		intervals:       intervals,
		trainingService: trainingService,
		config:          cfg,
	}
}

func (s *CombatServiceImpl) ResolveAttack(attacker *model.Character, target any) {
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
	if attacker.Stamina < 0 {
		attacker.Stamina = 0
	}
	s.messageService.UserService().GetConnection(attacker).Send(outgoing.NewUpdateUserStatsPacket(attacker))
}

func (s *CombatServiceImpl) resolvePVP(attacker *model.Character, victim *model.Character) {
	if s.mapService.IsSafeZone(attacker.Position) || s.mapService.IsSafeZone(victim.Position) ||
		!s.mapService.IsPkMap(attacker.Position.Map) || !s.mapService.IsPkMap(victim.Position.Map) {
		s.messageService.SendConsoleMessage(attacker, "No puedes combatir en zona segura.", outgoing.INFO)
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

		// Play miss sound
		s.messageService.SendToArea(&outgoing.PlayWavePacket{
			Wave: 2, // SND_MISS
			X:    victim.Position.X,
			Y:    victim.Position.Y,
		}, victim.Position)

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

	if damage < 1 {
		damage = 1
	}

	victim.Hp -= damage
	if victim.Hp < 0 {
		victim.Hp = 0
	}

	// Remove paralysis on hit
	if victim.Paralyzed || victim.Immobilized {
		victim.Paralyzed = false
		victim.Immobilized = false
		s.messageService.SendConsoleMessage(victim, "¡Has recuperado el movimiento!", outgoing.INFO)
	}

	// Feedback
	s.messageService.SendConsoleMessage(attacker, fmt.Sprintf("¡Has golpeado a %s por %d!", victim.Name, damage), outgoing.FIGHT)
	s.messageService.SendConsoleMessage(victim, fmt.Sprintf("¡%s te ha golpeado por %d!", attacker.Name, damage), outgoing.FIGHT)

	// Play hit sound
	s.messageService.SendToArea(&outgoing.PlayWavePacket{
		Wave: 10, // SND_HIT (Sword/Melee)
		X:    victim.Position.X,
		Y:    victim.Position.Y,
	}, victim.Position)

	// Blood FX
	s.messageService.SendToArea(&outgoing.CreateFxPacket{
		CharIndex: victim.CharIndex,
		FxID:      1, // Blood placeholder
		Loops:     0,
	}, victim.Position)

	if victim.Hp <= 0 {
		s.messageService.HandleDeath(victim, "")
	} else {
		connVictim := s.messageService.UserService().GetConnection(victim)
		if connVictim != nil {
			connVictim.Send(outgoing.NewUpdateUserStatsPacket(victim))
		}
	}
}

func (s *CombatServiceImpl) resolvePVE(attacker *model.Character, victim *model.WorldNPC) {
	if !victim.NPC.Hostile {
		s.messageService.SendConsoleMessage(attacker, "No puedes atacar a una criatura pacífica.", outgoing.INFO)
		return
	}

	if s.mapService.IsSafeZone(attacker.Position) || s.mapService.IsSafeZone(victim.Position) {
		s.messageService.SendConsoleMessage(attacker, "No puedes combatir en zona segura.", outgoing.INFO)
		return
	}

	weapon := s.getEquippedWeapon(attacker)

	// Hit check
	attackerPower := s.formulas.GetAttackPower(attacker, weapon)
	victimEvasion := victim.NPC.EvasionPower

	chance := s.formulas.CalculateHitChance(attackerPower, victimEvasion)

	if rand.Intn(100) >= chance {
		s.messageService.SendConsoleMessage(attacker, "¡Has fallado el golpe!", outgoing.FIGHT)

		// Play miss sound
		s.messageService.SendToArea(&outgoing.PlayWavePacket{
			Wave: 2, // SND_MISS
			X:    victim.Position.X,
			Y:    victim.Position.Y,
		}, victim.Position)

		return
	}

	// Damage
	damage := s.formulas.CalculateDamage(attacker, weapon, true)

	// NPC Defense
	damage -= victim.NPC.Defense

	if damage < 1 {
		damage = 1
	}

	victim.HP -= damage
	s.messageService.SendConsoleMessage(attacker, fmt.Sprintf("¡Has golpeado a la criatura por %d!", damage), outgoing.FIGHT)

	// Grant experience proportional to damage
	s.grantExperience(attacker, victim, damage)

	// Play hit sound
	s.messageService.SendToArea(&outgoing.PlayWavePacket{
		Wave: 10, // SND_HIT (Sword/Melee)
		X:    victim.Position.X,
		Y:    victim.Position.Y,
	}, victim.Position)

	if victim.HP <= 0 {
		s.handleNpcDeath(attacker, victim)
	}
}

func (s *CombatServiceImpl) NpcAtacaUser(npc *model.WorldNPC, victim *model.Character) bool {
	if victim.Dead {
		return false
	}

	if s.mapService.IsSafeZone(npc.Position) || s.mapService.IsSafeZone(victim.Position) {
		return false
	}

	// Check interval
	if !s.intervals.CanNPCAttack(npc) {
		return false
	}

	// Hit check
	attackerPower := npc.NPC.AttackPower
	victimEvasion := s.formulas.GetEvasionPower(victim)

	// Shield bonus
	if s.getEquippedShield(victim) != nil {
		victimEvasion += s.formulas.GetShieldEvasionPower(victim)
	}

	chance := s.formulas.CalculateHitChance(attackerPower, victimEvasion)

	if rand.Intn(100) >= chance {
		s.messageService.SendConsoleMessage(victim, fmt.Sprintf("¡%s ha fallado el golpe!", npc.NPC.Name), outgoing.FIGHT)

		// Play miss sound
		s.messageService.SendToArea(&outgoing.PlayWavePacket{
			Wave: 2, // SND_MISS
			X:    victim.Position.X,
			Y:    victim.Position.Y,
		}, victim.Position)

		// Update interval even on miss
		s.intervals.UpdateNPCLastAttack(npc)
		return true
	}

	// Damage calculation
	damage := utils.RandomNumber(npc.NPC.MinHit, npc.NPC.MaxHit)

	// Armor defense
	armor := s.getEquippedArmor(victim)
	if armor != nil {
		defense := utils.RandomNumber(armor.MinDef, armor.MaxDef)
		damage -= defense
	}

	if damage < 1 {
		damage = 1
	}

	victim.Hp -= damage
	if victim.Hp < 0 {
		victim.Hp = 0
	}

	// Remove paralysis on hit
	if victim.Paralyzed || victim.Immobilized {
		victim.Paralyzed = false
		victim.Immobilized = false
		s.messageService.SendConsoleMessage(victim, "¡Has recuperado el movimiento!", outgoing.INFO)
	}

	// Feedback
	s.messageService.SendConsoleMessage(victim, fmt.Sprintf("¡%s te ha golpeado por %d!", npc.NPC.Name, damage), outgoing.FIGHT)

	// Play hit sound
	s.messageService.SendToArea(&outgoing.PlayWavePacket{
		Wave: 10, // SND_HIT (Sword/Melee)
		X:    victim.Position.X,
		Y:    victim.Position.Y,
	}, victim.Position)

	// Blood FX
	s.messageService.SendToArea(&outgoing.CreateFxPacket{
		CharIndex: victim.CharIndex,
		FxID:      2, // Blood placeholder
		Loops:     0,
	}, victim.Position)

	if victim.Hp <= 0 {
		s.messageService.HandleDeath(victim, "")
	} else {
		connVictim := s.messageService.UserService().GetConnection(victim)
		if connVictim != nil {
			connVictim.Send(outgoing.NewUpdateUserStatsPacket(victim))
		}
	}

	// Update interval
	s.intervals.UpdateNPCLastAttack(npc)

	return true
}

func (s *CombatServiceImpl) grantExperience(attacker *model.Character, victim *model.WorldNPC, damage int) {
	if victim.NPC.MaxHp == 0 || victim.NPC.Exp == 0 || victim.RemainingExp <= 0 {
		return
	}

	baseShare := int(float64(damage) * (float64(victim.NPC.Exp) / float64(victim.NPC.MaxHp)))

	// Ensure at least 1 exp if damage was dealt and there's exp left
	if baseShare == 0 && damage > 0 && victim.RemainingExp > 0 {
		baseShare = 1
	}

	if baseShare > victim.RemainingExp {
		baseShare = victim.RemainingExp
	}

	if baseShare > 0 {
		expToGive := int(float64(baseShare) * s.config.XpMultiplier)
		attacker.Exp += expToGive
		victim.RemainingExp -= baseShare
		s.messageService.SendConsoleMessage(attacker, fmt.Sprintf("Has ganado %d puntos de experiencia.", expToGive), outgoing.FIGHT)
		s.trainingService.CheckLevel(attacker)
	}
}

func (s *CombatServiceImpl) handleNpcDeath(killer *model.Character, npc *model.WorldNPC) {
	s.messageService.SendToArea(&outgoing.CharacterRemovePacket{CharIndex: npc.Index}, npc.Position)

	if npc.RemainingExp > 0 {
		bonusExp := int(float64(npc.RemainingExp) * s.config.XpMultiplier)
		killer.Exp += bonusExp
		s.messageService.SendConsoleMessage(killer, fmt.Sprintf("¡Has matado a la criatura! Ganaste %d exp.", bonusExp), outgoing.INFO)
		npc.RemainingExp = 0
		s.trainingService.CheckLevel(killer)
	}

	// Drop logic
	for _, drop := range npc.NPC.Drops {
		obj := s.objectService.GetObject(drop.ObjectID)
		if obj != nil {
			amount := drop.Amount
			if obj.Type == model.OTMoney {
				amount = int(float64(amount) * s.config.GoldMultiplier)
			}
			worldObj := &model.WorldObject{
				Object: obj,
				Amount: amount,
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

	s.npcService.RemoveNPC(npc, s.mapService)
}

func (s *CombatServiceImpl) getEquippedWeapon(char *model.Character) *model.Object {
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

func (s *CombatServiceImpl) getEquippedShield(char *model.Character) *model.Object {
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

func (s *CombatServiceImpl) getEquippedArmor(char *model.Character) *model.Object {
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
