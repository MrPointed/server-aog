package service

import (
	"fmt"
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/persistence"
)

type NpcService struct {
	dao          *persistence.NpcDAO
	npcDefs      map[int]*model.NPC
	worldNpcs    map[int16]*model.WorldNPC
	indexManager *CharacterIndexManager
}

func NewNpcService(dao *persistence.NpcDAO, indexManager *CharacterIndexManager) *NpcService {
	return &NpcService{
		dao:          dao,
		npcDefs:      make(map[int]*model.NPC),
		worldNpcs:    make(map[int16]*model.WorldNPC),
		indexManager: indexManager,
	}
}

func (s *NpcService) LoadNpcs() error {
	fmt.Println("Loading NPCs from data file...")
	defs, err := s.dao.Load()
	if err != nil {
		return err
	}
	s.npcDefs = defs
	fmt.Printf("Successfully loaded %d NPC definitions.\n", len(s.npcDefs))
	return nil
}

func (s *NpcService) GetNpcDef(id int) *model.NPC {
	return s.npcDefs[id]
}

func (s *NpcService) SpawnNpc(id int, pos model.Position) *model.WorldNPC {
	def := s.GetNpcDef(id)
	if def == nil {
		return nil
	}

	worldNpc := &model.WorldNPC{
		NPC:          def,
		Position:     pos,
		Heading:      def.Heading,
		HP:           def.MaxHp,
		RemainingExp: def.Exp,
		Index:        s.indexManager.AssignIndex(),
	}

	s.worldNpcs[worldNpc.Index] = worldNpc
	return worldNpc
}

func (s *NpcService) GetWorldNpcs() map[int16]*model.WorldNPC {
	return s.worldNpcs
}

func (s *NpcService) MoveNpc(npc *model.WorldNPC, newPos model.Position, heading model.Heading, mapService *MapService, areaService *AreaService) {
	oldPos := npc.Position
	
	// Update MapService tiles
	m := mapService.GetMap(oldPos.Map)
	if m != nil {
		tile := m.GetTile(int(oldPos.X), int(oldPos.Y))
		if tile.NPC == npc {
			tile.NPC = nil
		}
	}

	m = mapService.GetMap(newPos.Map)
	if m != nil {
		tile := m.GetTile(int(newPos.X), int(newPos.Y))
		tile.NPC = npc
	}

	npc.Position = newPos
	npc.Heading = heading

	// Broadcast movement to nearby players
	areaService.NotifyNpcMovement(npc, oldPos)
}
