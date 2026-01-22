package service

import (
	"fmt"
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

	fmt.Printf("HandleUseSkillClick: Entered for User %s, Skill %d at %d,%d\n", user.Name, skill, x, y)

	// Basic validation

	if user.Dead {

		fmt.Println("HandleUseSkillClick: User is dead")

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

		fmt.Printf("HandleUseSkillClick: Default case hit for skill %d\n", skill)

		s.messageService.SendConsoleMessage(user, fmt.Sprintf("Skill %d no implementada en click.", skill), outgoing.INFO)

	}

}

func (s *SkillService) handleMagic(user *model.Character, x, y byte) {

	fmt.Printf("handleMagic: User %s casting spell %d at %d,%d\n", user.Name, user.SelectedSpell, x, y)

	if user.SelectedSpell == 0 {

		s.messageService.SendConsoleMessage(user, "Primero selecciona un hechizo.", outgoing.INFO)

		return

	}



	// Resolve target

	targetPos := model.Position{X: x, Y: y, Map: user.Position.Map}

	

	gameMap := s.mapService.GetMap(targetPos.Map)

	if gameMap == nil {

		fmt.Println("handleMagic: Map not found")

		return

	}



	tile := gameMap.GetTile(int(targetPos.X), int(targetPos.Y))

	var target any

	if tile.Character != nil {

		fmt.Printf("handleMagic: Found Character target %s\n", tile.Character.Name)

		target = tile.Character

	} else if tile.NPC != nil {

		fmt.Printf("handleMagic: Found NPC target %s\n", tile.NPC.NPC.Name)

		target = tile.NPC

	} else {

		// No entity found, treat as terrain target

		fmt.Println("handleMagic: No entity found, using position as target")

		target = targetPos

	}



	s.spellService.CastSpell(user, user.SelectedSpell, target)

	// Reset selected spell? Standard AO usually keeps it selected until changed or cast?

	// Usually keeps it.

	user.SelectedSpell = 0 // Some versions reset. Let's reset to force re-select/re-cast flow? 

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
