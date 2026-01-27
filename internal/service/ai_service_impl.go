package service

import (
	"math"
	"math/rand"
	"time"

	"github.com/ao-go-server/internal/model"
)

type AiServiceImpl struct {
	npcService     NpcService
	mapService     MapService
	areaService    AreaService
	userService    UserService
	combatService  CombatService
	messageService MessageService
	spellService   SpellService
	globalBalance  *model.GlobalBalanceConfig
	stopChan       chan struct{}
	ticks          uint64
	enabled        bool
}

func NewAiServiceImpl(npcService NpcService, mapService MapService, areaService AreaService, userService UserService, combatService CombatService, messageService MessageService, spellService SpellService, globalBalance *model.GlobalBalanceConfig) AiService {
	return &AiServiceImpl{
		npcService:     npcService,
		mapService:     mapService,
		areaService:    areaService,
		userService:    userService,
		combatService:  combatService,
		messageService: messageService,
		spellService:   spellService,
		globalBalance:  globalBalance,
		stopChan:       make(chan struct{}),
		enabled:        true,
	}
}

func (s *AiServiceImpl) Start() {
	s.enabled = true
	go s.aiLoop()
}

func (s *AiServiceImpl) Stop() {
	close(s.stopChan)
}

func (s *AiServiceImpl) SetEnabled(enabled bool) {
	s.enabled = enabled
}

func (s *AiServiceImpl) IsEnabled() bool {
	return s.enabled
}

func (s *AiServiceImpl) aiLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if s.enabled {
				s.ticks++
				s.processNpcs()
			}
		case <-s.stopChan:
			return
		}
	}
}

func (s *AiServiceImpl) processNpcs() {
	npcs := s.npcService.GetWorldNpcs()
	for _, npc := range npcs {
		s.handleNpcAI(npc)
	}
}

func (s *AiServiceImpl) handleNpcAI(npc *model.WorldNPC) {
	// 0. Handle Paralysis/Immobilization Expiration
	if (npc.Paralyzed || npc.Immobilized) && time.Since(npc.ParalyzedSince).Milliseconds() >= s.globalBalance.NPCParalizedTime {
		npc.Paralyzed = false
		npc.Immobilized = false
	}

	// 1. Hostility/Attack Logic - Handled by intervals in CombatService
	if npc.OwnerIndex == 0 {
		if npc.NPC.Type == model.NTGuard {
			s.guardiasAI(npc, false)
		} else if npc.NPC.Type == model.NTGuardCaos {
			s.guardiasAI(npc, true)
		} else if npc.NPC.Hostile {
			s.hostilMalvadoAI(npc)
		}
	}

	// 2. Movement Logic
	if npc.Paralyzed || npc.Immobilized || time.Since(npc.LastMovement).Milliseconds() < s.globalBalance.NPCIntervalMove {
		return
	}

	moved := false
	switch model.MovementType(npc.NPC.Movement) {
	case model.MovementRandom:
		if rand.Intn(15) == 3 {
			moved = s.moveRandomly(npc)
		}
		if !moved {
			if npc.NPC.Type == model.NTGuard {
				moved = s.persigueCriminal(npc)
			} else if npc.NPC.Type == model.NTGuardCaos {
				moved = s.persigueCiudadano(npc)
			}
		}

	case model.MovementHostile:
		moved = s.irUsuarioCercano(npc)

	case model.MovementDefense:
		moved = s.seguirAgresor(npc)

	case model.MovementGuardAttackCriminals:
		moved = s.persigueCriminal(npc)

	case model.MovementFollowOwner:
		moved = s.seguirAmo(npc)
		if !moved && rand.Intn(15) == 3 {
			moved = s.moveRandomly(npc)
		}

	case model.MovementObject:
		s.aiNpcObjeto(npc)
	}

	if moved {
		npc.LastMovement = time.Now()
	}
}

func (s *AiServiceImpl) moveRandomly(npc *model.WorldNPC) bool {
	heading := model.Heading(rand.Intn(4))
	return s.moveNpc(npc, heading)
}

func (s *AiServiceImpl) moveNpc(npc *model.WorldNPC, heading model.Heading) bool {
	newPos := npc.Position
	switch heading {
	case model.North:
		newPos.Y--
	case model.South:
		newPos.Y++
	case model.East:
		newPos.X++
	case model.West:
		newPos.X--
	}

	// Boundary checks
	if !s.mapService.IsInPlayableArea(int(newPos.X), int(newPos.Y)) {
		return false
	}

	//Valid position check
	if s.mapService.IsInvalidPosition(
		model.Position{
			X:   newPos.X,
			Y:   newPos.Y,
			Map: npc.Position.Map,
		}) {
		return false
	}

	// Check if tile is blocked
	gameMap := s.mapService.GetMap(newPos.Map)
	if gameMap == nil {
		return false
	}

	tile := gameMap.GetTile(int(newPos.X), int(newPos.Y))
	if tile.Blocked || tile.Character != nil || tile.NPC != nil {
		return false
	}

	// Water check: NPCs cannot walk on water unless there is a bridge
	hasBridge := tile.Layer2 > 0 || tile.Layer3 > 0
	if tile.IsWater && !hasBridge {
		return false
	}

	return s.npcService.MoveNpc(npc, newPos, heading, s.mapService, s.areaService)
}

func (s *AiServiceImpl) guardiasAI(npc *model.WorldNPC, delCaos bool) {
	// Look in 4 directions for targets
	for h := model.North; h <= model.West; h++ {
		targetPos := npc.Position
		switch h {
		case model.North:
			targetPos.Y--
		case model.East:
			targetPos.X++
		case model.South:
			targetPos.Y++
		case model.West:
			targetPos.X--
		}

		if npc.Immobilized && h != npc.Heading {
			continue
		}

		gameMap := s.mapService.GetMap(targetPos.Map)
		if gameMap == nil {
			continue
		}
		tile := gameMap.GetTile(int(targetPos.X), int(targetPos.Y))
		if tile.Character != nil {
			victim := tile.Character
			if !victim.Dead {
				isCriminal := victim.Faccion.Criminal
				if !delCaos {
					if isCriminal || npc.AttackedBy == victim.Name {
						if s.npcAtacaUser(npc, victim) {
							return
						}
					}
				} else {
					if !isCriminal || npc.AttackedBy == victim.Name {
						if s.npcAtacaUser(npc, victim) {
							return
						}
					}
				}
			}
		}
	}
}

func (s *AiServiceImpl) hostilMalvadoAI(npc *model.WorldNPC) {
	for h := model.North; h <= model.West; h++ {
		targetPos := npc.Position
		switch h {
		case model.North:
			targetPos.Y--
		case model.East:
			targetPos.X++
		case model.South:
			targetPos.Y++
		case model.West:
			targetPos.X--
		}

		if npc.Immobilized && h != npc.Heading {
			continue
		}

		gameMap := s.mapService.GetMap(targetPos.Map)
		if gameMap == nil {
			continue
		}
		tile := gameMap.GetTile(int(targetPos.X), int(targetPos.Y))
		if tile.Character != nil {
			victim := tile.Character
			if !victim.Dead {
				if s.npcAtacaUser(npc, victim) {
					return
				}
			}
		}
	}
}

func (s *AiServiceImpl) irUsuarioCercano(npc *model.WorldNPC) bool {
	if npc.Immobilized {
		// Just attack if someone is in range
		s.hostilMalvadoAI(npc)
		return false
	}

	// Find closest user in range
	var closestUser *model.Character
	minDist := 15 // Range of vision

	for _, user := range s.userService.GetLoggedCharacters() {
		if user.Position.Map != npc.Position.Map || user.Dead {
			continue
		}
		dist := npc.Position.GetDistance(user.Position)
		if dist <= minDist {
			closestUser = user
			minDist = dist
		}
	}

	if closestUser != nil {
		if minDist <= 1 {
			if s.npcAtacaUser(npc, closestUser) {
				return false // Attacking isn't moving
			}
		} else {
			heading := s.findDirection(npc.Position, closestUser.Position)
			return s.moveNpc(npc, heading)
		}
	} else if rand.Intn(10) == 0 {
		return s.moveRandomly(npc)
	}
	return false
}

func (s *AiServiceImpl) persigueCriminal(npc *model.WorldNPC) bool {
	if npc.Immobilized {
		return false
	}

	var target *model.Character
	minDist := 15

	for _, user := range s.userService.GetLoggedCharacters() {
		if user.Position.Map != npc.Position.Map || user.Dead || !user.Faccion.Criminal {
			continue
		}
		dist := npc.Position.GetDistance(user.Position)
		if dist <= minDist {
			target = user
			minDist = dist
		}
	}

	if target != nil {
		heading := s.findDirection(npc.Position, target.Position)
		return s.moveNpc(npc, heading)
	}
	return false
}

func (s *AiServiceImpl) persigueCiudadano(npc *model.WorldNPC) bool {
	if npc.Immobilized {
		return false
	}

	var target *model.Character
	minDist := 15

	for _, user := range s.userService.GetLoggedCharacters() {
		if user.Position.Map != npc.Position.Map || user.Dead || user.Faccion.Criminal {
			continue
		}
		dist := npc.Position.GetDistance(user.Position)
		if dist <= minDist {
			target = user
			minDist = dist
		}
	}

	if target != nil {
		heading := s.findDirection(npc.Position, target.Position)
		return s.moveNpc(npc, heading)
	}
	return false
}

func (s *AiServiceImpl) seguirAgresor(npc *model.WorldNPC) bool {
	if npc.AttackedBy == "" {
		return false
	}

	var target *model.Character
	for _, user := range s.userService.GetLoggedCharacters() {
		if user.Name == npc.AttackedBy {
			target = user
			break
		}
	}

	if target == nil || target.Position.Map != npc.Position.Map || target.Dead {
		npc.AttackedBy = ""
		return false
	}

	dist := npc.Position.GetDistance(target.Position)
	if dist <= 1 {
		if s.npcAtacaUser(npc, target) {
			return false
		}
	} else if !npc.Immobilized {
		heading := s.findDirection(npc.Position, target.Position)
		return s.moveNpc(npc, heading)
	}
	return false
}

func (s *AiServiceImpl) seguirAmo(npc *model.WorldNPC) bool {
	owner := s.userService.GetCharacterByIndex(int16(npc.OwnerIndex))
	if owner == nil || owner.Position.Map != npc.Position.Map {
		return false
	}

	dist := npc.Position.GetDistance(owner.Position)
	if dist > 3 && dist < 15 {
		heading := s.findDirection(npc.Position, owner.Position)
		return s.moveNpc(npc, heading)
	}
	return false
}

func (s *AiServiceImpl) aiNpcObjeto(npc *model.WorldNPC) {
	// NPC objects don't move, they just attack nearby users
	s.hostilMalvadoAI(npc)
}

func (s *AiServiceImpl) npcAtacaUser(npc *model.WorldNPC, victim *model.Character) bool {
	if npc.Paralyzed {
		return false
	}

	// Head to player before hitting
	heading := s.findDirection(npc.Position, victim.Position)
	if npc.Heading != heading {
		s.npcService.ChangeNpcHeading(npc, heading, s.areaService)
	}

	if npc.NPC.CastsSpells > 0 {
		if npc.NPC.DoubleAttack {
			if rand.Intn(2) == 0 {
				return s.combatService.NpcAtacaUser(npc, victim)
			}
		}
		if s.npcLanzaUnSpell(npc, victim) {
			return true
		}
	}
	return s.combatService.NpcAtacaUser(npc, victim)
}

func (s *AiServiceImpl) npcLanzaUnSpell(npc *model.WorldNPC, victim *model.Character) bool {
	if victim.Invisible || victim.Hidden {
		return false
	}

	// Heading is already set by npcAtacaUser
	if npc.NPC.CastsSpells > 0 && len(npc.NPC.Spells) > 0 {
		idx := rand.Intn(len(npc.NPC.Spells))
		spellID := npc.NPC.Spells[idx]
		return s.spellService.NpcLanzaSpellSobreUser(npc, victim, spellID)
	}
	return false
}

func (s *AiServiceImpl) findDirection(from, to model.Position) model.Heading {
	dx := int(to.X) - int(from.X)
	dy := int(to.Y) - int(from.Y)

	if math.Abs(float64(dx)) > math.Abs(float64(dy)) {
		if dx > 0 {
			return model.East
		}
		return model.West
	} else {
		if dy > 0 {
			return model.South
		}
		return model.North
	}
}
