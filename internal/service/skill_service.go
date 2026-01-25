package service

import (
	"log/slog"
	"math/rand"

	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type SkillService struct {

	mapService     *MapService

	objectService  *ObjectService

	messageService *MessageService

	userService    *UserService

	npcService     *NpcService

	spellService   *SpellService

	intervals      *IntervalService

}



func NewSkillService(mapService *MapService, objectService *ObjectService, messageService *MessageService, userService *UserService, npcService *NpcService, spellService *SpellService, intervals *IntervalService) *SkillService {

	return &SkillService{

		mapService:		mapService,

		objectService:	objectService,

		messageService:	messageService,

		userService:	userService,

		npcService:		npcService,

		spellService:	spellService,

		intervals:		intervals,

	}

}



func (s *SkillService) HandleUseSkillClick(user *model.Character, skill model.Skill, x, y byte) {

	slog.Debug("HandleUseSkillClick", "user", user.Name, "skill", skill, "x", x, "y", y)

	// Basic validation

	if user.Dead {

		slog.Debug("HandleUseSkillClick: User is dead")

		return

	}



	// Check intervals for specific working skills (Mining, Fishing, Lumber, Stealing, Taming)

	if skill != model.Magic {

		if !s.intervals.CanWork(user) {

			return

		}

	}

	

	// Range check (using Manhattan distance as per VB6)

	dist := int(user.Position.X) - int(x)

	if dist < 0 {

		dist = -dist

	}

	dy := int(user.Position.Y) - int(y)

	if dy < 0 {

		dy = -dy

	}

	if dist+dy > 2 { // RANGO_VISION_X check? VB6 uses simple distance check for some skills, vision for others.

		// VB6: If Not InRangoVision -> WritePosUpdate.

		// Here we assume vision check passed or irrelevant for now.

	}



	switch skill {

	case model.Magic:

		s.handleMagic(user, x, y)



	case model.Fishing:

		s.handleFishing(user, x, y)

		s.intervals.UpdateLastWork(user)



	case model.Steal:

		s.handleStealing(user, x, y)

		s.intervals.UpdateLastWork(user)



	case model.Lumber:

		s.handleLumber(user, x, y)

		s.intervals.UpdateLastWork(user)



	case model.Mining:

		s.handleMining(user, x, y)

		s.intervals.UpdateLastWork(user)



	case model.Tame:

		s.handleTaming(user, x, y)

		s.intervals.UpdateLastWork(user)



		default:



	



			slog.Debug("HandleUseSkillClick: Default case hit", "skill", skill)



	



		}

}

func (s *SkillService) handleMagic(user *model.Character, x, y byte) {

	slog.Debug("handleMagic", "user", user.Name, "spell", user.SelectedSpell, "x", x, "y", y)

	if user.SelectedSpell == 0 {

		return

	}





		m := s.mapService.GetMap(user.Position.Map)





	





		if m == nil {





	





			slog.Debug("handleMagic: Map not found")





	





			return





	





		}





	





	





	





		tile := m.GetTile(int(x), int(y))





	





	





	





		if tile.Character != nil {





	





			slog.Debug("handleMagic: Found Character target", "name", tile.Character.Name)





	





			s.spellService.CastSpell(user, user.SelectedSpell, tile.Character)





	





		} else if tile.NPC != nil {





	





			slog.Debug("handleMagic: Found NPC target", "name", tile.NPC.NPC.Name)





	





			s.spellService.CastSpell(user, user.SelectedSpell, tile.NPC)





	





		} else {





	





			slog.Debug("handleMagic: No entity found, using position as target")





	





			targetPos := model.Position{X: x, Y: y, Map: user.Position.Map}





	





			s.spellService.CastSpell(user, user.SelectedSpell, targetPos)





	





		}





	}

func (s *SkillService) handleFishing(user *model.Character, x, y byte) {
	// 1. Check Tool
	// 2. Check Water
	// 3. Fish
	s.messageService.SendConsoleMessage(user, "Pesca no implementada aún.", outgoing.INFO)
}

func (s *SkillService) handleStealing(user *model.Character, x, y byte) {
	// 1. Check Target User
	// 2. Check Safe Zone
	// 3. Steal
	s.messageService.SendConsoleMessage(user, "Robar no implementado aún.", outgoing.INFO)
}

func (s *SkillService) handleLumber(user *model.Character, x, y byte) {
	// 1. Check Tool (Axe)
	// 2. Check Tree
	// 3. Get Wood
	s.messageService.SendConsoleMessage(user, "Talar no implementado aún.", outgoing.INFO)
}

func (s *SkillService) handleMining(user *model.Character, x, y byte) {
	// 1. Check Tool (Pickaxe)
	// 2. Check Deposit
	// 3. Get Ore
	s.messageService.SendConsoleMessage(user, "Minería no implementada aún.", outgoing.INFO)
}

func (s *SkillService) handleTaming(user *model.Character, x, y byte) {
	// 1. Check Target NPC
	// 2. Check if Tameable
	// 3. Tame

	// Logic from VB6:
	// Call LookatTile -> get tN (Target NPC index)
	// If tN > 0 -> Check Domable -> Check Distance -> DoDomar

	// We need to find NPC at x,y
	targetPos := model.Position{X: x, Y: y, Map: user.Position.Map}
	npc := s.mapService.GetNPCAt(targetPos)

	if npc == nil {
		s.messageService.SendConsoleMessage(user, "No hay ninguna criatura allí.", outgoing.INFO)
		return
	}

	// Distance check (already done in main handler mostly, but VB6 does it again: > 2 blocks)
	dist := user.Position.GetDistance(targetPos)
	if dist > 2 {
		s.messageService.SendConsoleMessage(user, "Estás demasiado lejos.", outgoing.INFO)
		return
	}

	// TODO: Check if NPC is tameable (Need 'Domable' field in NPC model)
	// TODO: Check if NPC is attacked by user

	// Success check based on Skill + Charisma + Level vs NPC difficulty
	success := rand.Float32() > 0.5 // Placeholder

	if success {
		s.messageService.SendConsoleMessage(user, "¡Has domado la criatura!", outgoing.INFO)
		// TODO: Convert NPC to Pet
	} else {
		s.messageService.SendConsoleMessage(user, "Has fallado en el intento.", outgoing.INFO)
	}
}
