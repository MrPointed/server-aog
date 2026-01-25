package service

import (
	"log/slog"
	"time"
)

type ResourceManager struct {
	objectService *ObjectService
	npcService    *NpcService
	mapService    *MapService
	spellService  *SpellService
	cityService   *CityService
}

func NewResourceManager(objectService *ObjectService, npcService *NpcService, mapService *MapService, spellService *SpellService, cityService *CityService) *ResourceManager {
	return &ResourceManager{
		objectService: objectService,
		npcService:    npcService,
		mapService:    mapService,
		spellService:  spellService,
		cityService:   cityService,
	}
}

func (s *ResourceManager) LoadAll() {
	start := time.Now()
	slog.Info("Starting resource loading...")

	if err := s.objectService.LoadObjects(); err != nil {
		slog.Error("Error loading objects", "error", err)
	}
	if err := s.npcService.LoadNpcs(); err != nil {
		slog.Error("Error loading NPCs", "error", err)
	}

	if err := s.cityService.LoadCities(); err != nil {
		slog.Error("Error loading cities", "error", err)
	}
	if err := s.spellService.LoadSpells(); err != nil {
		slog.Error("Error loading spells", "error", err)
	}

	// Maps are the slowest and depend on Objects and NPCs
	s.mapService.LoadMaps()

	slog.Info("All resources loaded", "duration", time.Since(start))
}
