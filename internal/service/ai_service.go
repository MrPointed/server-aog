package service

import (
	"math/rand"
	"time"

	"github.com/ao-go-server/internal/model"
)

type AIService struct {
	npcService  *NpcService
	mapService  *MapService
	areaService *AreaService
	stopChan    chan struct{}
}

func NewAIService(npcService *NpcService, mapService *MapService, areaService *AreaService) *AIService {
	return &AIService{
		npcService:  npcService,
		mapService:  mapService,
		areaService: areaService,
		stopChan:    make(chan struct{}),
	}
}

func (s *AIService) Start() {
	go s.aiLoop()
}

func (s *AIService) Stop() {
	close(s.stopChan)
}

func (s *AIService) aiLoop() {
	ticker := time.NewTicker(3 * time.Second)
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
		if npc.NPC.Movement == 1 {
			continue
		}

		// Random chance to move
		if rand.Float32() > 0.3 {
			continue
		}

		// Random direction
		heading := model.Heading(rand.Intn(4))

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
			continue
		}

		// Check if tile is blocked
		gameMap := s.mapService.GetMap(newPos.Map)
		if gameMap == nil {
			continue
		}

		tile := gameMap.GetTile(int(newPos.X), int(newPos.Y))
		if tile.Blocked || tile.Character != nil || tile.NPC != nil {
			continue
		}

		// Move it
		s.npcService.MoveNpc(npc, newPos, heading, s.mapService, s.areaService)
	}
}
