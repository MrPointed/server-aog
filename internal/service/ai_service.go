package service

import (
	"math"
	"math/rand"
	"time"

	"github.com/ao-go-server/internal/model"
)

type AIService struct {
	npcService     *NpcService
	mapService     *MapService
	areaService    *AreaService
	userService    *UserService
	combatService  *CombatService
	messageService *MessageService
	spellService   *SpellService
	stopChan       chan struct{}
}

func NewAIService(npcService *NpcService, mapService *MapService, areaService *AreaService, userService *UserService, combatService *CombatService, messageService *MessageService, spellService *SpellService) *AIService {
	return &AIService{
		npcService:     npcService,
		mapService:     mapService,
		areaService:    areaService,
		userService:    userService,
		combatService:  combatService,
		messageService: messageService,
		spellService:   spellService,
		stopChan:       make(chan struct{}),
	}
}

func (s *AIService) Start() {
	go s.aiLoop()
}

func (s *AIService) Stop() {
	close(s.stopChan)
}

func (s *AIService) aiLoop() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.processNpcs()
		case <-s.stopChan:
			return
		}
	}
}

func (s *AIService) processNpcs() {
	npcs := s.npcService.GetWorldNpcs()
	for _, npc := range npcs {
		s.handleNpcAI(npc)
	}
}

func (s *AIService) handleNpcAI(npc *model.WorldNPC) {
	// 1. Hostility/Attack Logic
	if npc.MaestroUser == 0 {
		if npc.NPC.Type == model.NTGuard {
			s.guardiasAI(npc, false)
		} else if npc.NPC.Type == model.NTGuardCaos {
			s.guardiasAI(npc, true)
		} else if npc.NPC.Hostile {
			s.hostilMalvadoAI(npc)
		}
	}

	// 2. Movement Logic
	switch model.MovementType(npc.NPC.Movement) {
	case model.MovementRandom:
		if npc.Inmovilizado {
			return
		}
		if rand.Intn(12) == 3 {
			s.moveRandomly(npc)
		}
		if npc.NPC.Type == model.NTGuard {
			s.persigueCriminal(npc)
		} else if npc.NPC.Type == model.NTGuardCaos {
			s.persigueCiudadano(npc)
		}

	case model.MovementHostile:
		s.irUsuarioCercano(npc)

	case model.MovementDefense:
		s.seguirAgresor(npc)

	case model.MovementGuardAttackCriminals:
		s.persigueCriminal(npc)

	case model.MovementFollowOwner:
		if npc.Inmovilizado {
			return
		}
		s.seguirAmo(npc)
		if rand.Intn(12) == 3 {
			s.moveRandomly(npc)
		}

	case model.MovementObject:
		s.aiNpcObjeto(npc)
	}
}

func (s *AIService) moveRandomly(npc *model.WorldNPC) {
	heading := model.Heading(rand.Intn(4))
	s.moveNpc(npc, heading)
}

func (s *AIService) moveNpc(npc *model.WorldNPC, heading model.Heading) bool {
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
	if newPos.X < 1 || newPos.X >= 100 || newPos.Y < 1 || newPos.Y >= 100 {
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

	s.npcService.MoveNpc(npc, newPos, heading, s.mapService, s.areaService)
	return true
}

func (s *AIService) guardiasAI(npc *model.WorldNPC, delCaos bool) {
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

		if npc.Inmovilizado && h != npc.Heading {
			continue
		}

		tile := s.mapService.GetMap(targetPos.Map).GetTile(int(targetPos.X), int(targetPos.Y))
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

func (s *AIService) hostilMalvadoAI(npc *model.WorldNPC) {
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

		if npc.Inmovilizado && h != npc.Heading {
			continue
		}

		tile := s.mapService.GetMap(targetPos.Map).GetTile(int(targetPos.X), int(targetPos.Y))
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

func (s *AIService) irUsuarioCercano(npc *model.WorldNPC) {
	if npc.Inmovilizado {
		// Just attack if someone is in range
		s.hostilMalvadoAI(npc)
		return
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
				return
			}
		} else {
			heading := s.findDirection(npc.Position, closestUser.Position)
			s.moveNpc(npc, heading)
		}
	} else if rand.Intn(10) == 0 {
		s.moveRandomly(npc)
	}
}

func (s *AIService) persigueCriminal(npc *model.WorldNPC) {
	if npc.Inmovilizado {
		return
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
		s.moveNpc(npc, heading)
	}
}

func (s *AIService) persigueCiudadano(npc *model.WorldNPC) {
	if npc.Inmovilizado {
		return
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
		s.moveNpc(npc, heading)
	}
}

func (s *AIService) seguirAgresor(npc *model.WorldNPC) {
	if npc.AttackedBy == "" {
		return
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
		return
	}

	dist := npc.Position.GetDistance(target.Position)
	if dist <= 1 {
		if s.npcAtacaUser(npc, target) {
			return
		}
	} else if !npc.Inmovilizado {
		heading := s.findDirection(npc.Position, target.Position)
		s.moveNpc(npc, heading)
	}
}

func (s *AIService) seguirAmo(npc *model.WorldNPC) {
	owner := s.userService.GetCharacterByIndex(int16(npc.MaestroUser))
	if owner == nil || owner.Position.Map != npc.Position.Map {
		return
	}

	dist := npc.Position.GetDistance(owner.Position)
	if dist > 3 && dist < 15 {
		heading := s.findDirection(npc.Position, owner.Position)
		s.moveNpc(npc, heading)
	}
}

func (s *AIService) aiNpcObjeto(npc *model.WorldNPC) {
	// NPC objects don't move, they just attack nearby users
	s.hostilMalvadoAI(npc)
}

func (s *AIService) npcAtacaUser(npc *model.WorldNPC, victim *model.Character) bool {
	// Head to player before hitting
	heading := s.findDirection(npc.Position, victim.Position)
	if npc.Heading != heading {
		s.npcService.ChangeNpcHeading(npc, heading, s.areaService)
	}

	if npc.NPC.LanzaSpells > 0 {
		if npc.NPC.AtacaDoble {
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

func (s *AIService) npcLanzaUnSpell(npc *model.WorldNPC, victim *model.Character) bool {
	if victim.Invisible || victim.Hidden {
		return false
	}

	// Heading is already set by npcAtacaUser
	if npc.NPC.LanzaSpells > 0 && len(npc.NPC.Spells) > 0 {
		idx := rand.Intn(len(npc.NPC.Spells))
		spellID := npc.NPC.Spells[idx]
		return s.spellService.NpcLanzaSpellSobreUser(npc, victim, spellID)
	}
	return false
}

func (s *AIService) findDirection(from, to model.Position) model.Heading {
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

