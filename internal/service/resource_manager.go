package service

import (
	"fmt"
	"time"
)

type ResourceManager struct {
	objectService *ObjectService
	npcService    *NpcService
	mapService     *MapService
	spellService   *SpellService
	cityService    *CityService
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

func (rm *ResourceManager) LoadAll() {
	start := time.Now()
	fmt.Println("Starting resource loading...")

	// Objects and NPCs must be loaded before maps
	if err := rm.objectService.LoadObjects(); err != nil {
		fmt.Printf("Error loading objects: %v\n", err)
	}
	if err := rm.npcService.LoadNpcs(); err != nil {
		fmt.Printf("Error loading NPCs: %v\n", err)
	}

	// These can be loaded concurrently with each other or just sequentially since they are fast
	if err := rm.cityService.LoadCities(); err != nil {
		fmt.Printf("Error loading cities: %v\n", err)
	}
	if err := rm.spellService.LoadSpells(); err != nil {
		fmt.Printf("Error loading spells: %v\n", err)
	}

	// Maps are the slowest and depend on Objects and NPCs
	rm.mapService.LoadMapsConcurrent()

	fmt.Printf("All resources loaded in %v\n", time.Since(start))
}
