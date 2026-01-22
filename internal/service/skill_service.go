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
}

func NewSkillService(mapService *MapService, objectService *ObjectService, messageService *MessageService, userService *UserService, npcService *NpcService, spellService *SpellService) *SkillService {
	return &SkillService{
		mapService:     mapService,
		objectService:  objectService,
		messageService: messageService,
		userService:    userService,
		npcService:     npcService,
		spellService:   spellService,
	}
}

func (s *SkillService) HandleUseSkillClick(user *model.Character, skill model.Skill, x, y byte) {
	// Basic validation
	if user.Dead {
		return
	}
	
	// Range check (using Manhattan distance as per VB6)
	dist := int(user.Position.X) - int(x)
	if dist < 0 { dist = -dist }
	dy := int(user.Position.Y) - int(y)
	if dy < 0 { dy = -dy }
	if dist + dy > 2 { // RANGO_VISION_X check? VB6 uses simple distance check for some skills, vision for others.
		// VB6: If Not InRangoVision -> WritePosUpdate.
		// Here we assume vision check passed or irrelevant for now.
	}

	switch skill {
	case model.Magic:
		s.handleMagic(user, x, y)
	
	case model.Fishing:
		s.handleFishing(user, x, y)

	case model.Steal:
		s.handleStealing(user, x, y)

	case model.Lumber:
		s.handleLumber(user, x, y)

	case model.Mining:
		s.handleMining(user, x, y)

	case model.Tame:
		s.handleTaming(user, x, y)

	default:
		s.messageService.SendConsoleMessage(user, fmt.Sprintf("Skill %d no implementada en click.", skill), outgoing.INFO)
	}
}

func (s *SkillService) handleMagic(user *model.Character, x, y byte) {
	if user.SelectedSpell == 0 {
		s.messageService.SendConsoleMessage(user, "Primero selecciona un hechizo.", outgoing.INFO)
		return
	}

	// Resolve target
	targetPos := model.Position{X: x, Y: y, Map: user.Position.Map}
	
	gameMap := s.mapService.GetMap(targetPos.Map)
	if gameMap == nil {
		return
	}

	tile := gameMap.GetTile(int(targetPos.X), int(targetPos.Y))
	var target any
	if tile.Character != nil {
		target = tile.Character
	} else if tile.NPC != nil {
		target = tile.NPC
	} else {
		// Target is the tile/ground/self?
		// If clicking on self, tile.Character should be self.
		// If clicking on ground, maybe area spell?
		// For now, if no target found, do nothing or Self?
		// CastSpell expects a target.
		// If tile.Character is nil, maybe user clicked empty tile.
		// Standard AO CastSpell usually requires a valid target (User or NPC) unless it's a specific spell type.
		// Let's pass nil target and let SpellService handle it (or self?).
		// Actually SpellService.CastSpell expects 'target any'.
		// If I pass nil, it might crash or fail.
		// Let's try casting on Self if coordinates match User?
		// But tile.Character handles that.
		
		// If nothing there, maybe cancel?
		s.messageService.SendConsoleMessage(user, "No hay objetivo.", outgoing.INFO)
		return
	}

	s.spellService.CastSpell(user, user.SelectedSpell, target)
	// Reset selected spell? Standard AO usually keeps it selected until changed or cast?
	// Usually keeps it.
	user.SelectedSpell = 0 // Some versions reset. Let's reset to force re-select/re-cast flow? 
	// Actually, standard behavior: Click "Lanzar" -> Cursor -> Click -> Cast -> Cursor gone. Spell still selected in list?
	// Yes, but 'flags.Hechizo' is usually reset.
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
