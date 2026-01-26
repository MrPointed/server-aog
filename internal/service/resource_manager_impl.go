package service

import (
	"log/slog"
	"time"
)

type ResourceManagerImpl struct {
	objectService ObjectService
	npcService    NpcService
	mapService    MapService
	spellService  SpellService
	cityService   CityService
}

func NewResourceManagerImpl(objectService ObjectService, npcService NpcService, mapService MapService, spellService SpellService, cityService CityService) ResourceManager {
	return &ResourceManagerImpl{
		objectService: objectService,
		npcService:    npcService,
		mapService:    mapService,
		spellService:  spellService,
		cityService:   cityService,
	}
}

func (s *ResourceManagerImpl) LoadAll() {
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
