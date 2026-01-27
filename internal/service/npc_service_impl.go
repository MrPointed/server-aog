package service

import (
	"log/slog"
	"sync"

	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/persistence"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type NpcServiceImpl struct {
	dao          persistence.NpcRepository
	npcDefs      map[int]*model.NPC
	worldNpcs    map[int16]*model.WorldNPC
	indexManager *CharacterIndexManager
	mu           sync.RWMutex
}

func NewNpcServiceImpl(dao persistence.NpcRepository, indexManager *CharacterIndexManager) NpcService {
	return &NpcServiceImpl{
		dao:          dao,
		npcDefs:      make(map[int]*model.NPC),
		worldNpcs:    make(map[int16]*model.WorldNPC),
		indexManager: indexManager,
	}
}

func (s *NpcServiceImpl) LoadNpcs() error {
	defs, err := s.dao.Load()
	if err != nil {
		return err
	}
	s.npcDefs = defs
	slog.Info("Successfully loaded NPCs", "count", len(s.npcDefs))
	return nil
}

func (s *NpcServiceImpl) GetNpcDef(id int) *model.NPC {
	return s.npcDefs[id]
}

func (s *NpcServiceImpl) SpawnNpc(id int, pos model.Position) *model.WorldNPC {
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
		Respawn:      def.Respawn,
	}

	s.mu.Lock()
	s.worldNpcs[worldNpc.Index] = worldNpc
	s.mu.Unlock()
	return worldNpc
}

func (s *NpcServiceImpl) RemoveNPC(npc *model.WorldNPC, mapService MapService) {
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

func (s *NpcServiceImpl) GetWorldNpcs() []*model.WorldNPC {
	s.mu.RLock()
	defer s.mu.RUnlock()

	npcs := make([]*model.WorldNPC, 0, len(s.worldNpcs))
	for _, npc := range s.worldNpcs {
		npcs = append(npcs, npc)
	}
	return npcs
}

func (s *NpcServiceImpl) GetWorldNpcByIndex(index int16) *model.WorldNPC {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.worldNpcs[index]
}

func (s *NpcServiceImpl) ChangeNpcHeading(npc *model.WorldNPC, heading model.Heading, areaService AreaService) {
	npc.Heading = heading
	areaService.BroadcastToArea(npc.Position, &outgoing.NpcChangePacket{Npc: npc})
}

func (s *NpcServiceImpl) MoveNpc(npc *model.WorldNPC, newPos model.Position, heading model.Heading, mapService MapService, areaService AreaService) bool {
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
