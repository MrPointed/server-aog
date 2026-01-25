package service

import (
	"log/slog"
	"sync"

	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/persistence"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type NpcService struct {
	dao          *persistence.NpcDAO
	npcDefs      map[int]*model.NPC
	worldNpcs    map[int16]*model.WorldNPC
	indexManager *CharacterIndexManager
	mu           sync.RWMutex
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
	defs, err := s.dao.Load()
	if err != nil {
		return err
	}
	s.npcDefs = defs
	slog.Info("Successfully loaded NPCs", "count", len(s.npcDefs))
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
		Respawn:      def.ReSpawn,
	}

	s.mu.Lock()
	s.worldNpcs[worldNpc.Index] = worldNpc
	s.mu.Unlock()
	return worldNpc
}

func (s *NpcService) RemoveNPC(npc *model.WorldNPC, mapService *MapService) {
	s.mu.Lock()
	delete(s.worldNpcs, npc.Index)
	s.mu.Unlock()

	s.indexManager.FreeIndex(npc.Index)

	// Remove from map
	mapService.RemoveNPC(npc)

	// Respawn logic
	if npc.Respawn {
		mapService.SpawnNpcInMap(npc.NPC.ID, npc.Position.Map)
	}
}

func (s *NpcService) GetWorldNpcs() []*model.WorldNPC {
	s.mu.RLock()
	defer s.mu.RUnlock()

	npcs := make([]*model.WorldNPC, 0, len(s.worldNpcs))
	for _, npc := range s.worldNpcs {
		npcs = append(npcs, npc)
	}
	return npcs
}

func (s *NpcService) GetWorldNpcByIndex(index int16) *model.WorldNPC {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.worldNpcs[index]
}

func (s *NpcService) ChangeNpcHeading(npc *model.WorldNPC, heading model.Heading, areaService *AreaService) {
	npc.Heading = heading
	areaService.BroadcastToArea(npc.Position, &outgoing.NpcChangePacket{Npc: npc})
}

func (s *NpcService) MoveNpc(npc *model.WorldNPC, newPos model.Position, heading model.Heading, mapService *MapService, areaService *AreaService) bool {
	oldPos := npc.Position

	if !mapService.MoveNpc(npc, newPos) {
		return false
	}

	npc.Position = newPos
	npc.Heading = heading

	// Broadcast movement to nearby players
	areaService.NotifyNpcMovement(npc, oldPos)
	return true
}
