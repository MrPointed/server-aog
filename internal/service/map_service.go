package service

import (
	"encoding/gob"
	"fmt"
	"math/rand"
	"os"
	"sync"

	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/persistence"
)

const MapCacheFile = "resources/maps.cache"

type MapService struct {
	mapDAO        *persistence.MapDAO
	objectService *ObjectService
	npcService    *NpcService
	maps          map[int]*model.Map
}

func NewMapService(mapDAO *persistence.MapDAO, objectService *ObjectService, npcService *NpcService) *MapService {
	// Register types for gob
	gob.Register(model.Position{})
	return &MapService{
		mapDAO:        mapDAO,
		objectService: objectService,
		npcService:    npcService,
		maps:          make(map[int]*model.Map),
	}
}

func (s *MapService) LoadMaps() {
	fmt.Println("Loading maps...")
	maps, err := s.mapDAO.Load()
	if err != nil {
		fmt.Printf("Error loading maps: %v\n", err)
		return
	}
	for _, m := range maps {
		m.Characters = make(map[int16]*model.Character)
		m.Npcs = make(map[int16]*model.WorldNPC)
		s.resolveMapEntities(m)
		s.maps[m.Id] = m
	}
	fmt.Printf("Loaded %d maps\n", len(s.maps))
}

func (s *MapService) LoadMapsConcurrent() {
	if s.LoadCache() {
		// We still need to resolve entities because pointers to Object/NPC definitions
		// are not cached and depend on the current objects.dat/npcs.dat
		for _, m := range s.maps {
			s.resolveMapEntities(m)
		}

		fmt.Println("Successfully loaded maps.")
		return
	}

	amount := s.mapDAO.GetMapsAmount()
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 1; i <= amount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			m, err := s.mapDAO.LoadMap(id)
			if err != nil {
				fmt.Printf("Error loading map %d: %v\n", id, err)
				return
			}
			m.Characters = make(map[int16]*model.Character)
			m.Npcs = make(map[int16]*model.WorldNPC)

			mu.Lock()
			s.maps[m.Id] = m
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	// After loading all, we save the cache (before resolving entities to keep it clean)
	s.SaveCache()

	// Now resolve entities
	for _, m := range s.maps {
		s.resolveMapEntities(m)
	}

	fmt.Printf("Loaded %d maps\n", len(s.maps))
}

func (s *MapService) SaveCache() {
	file, err := os.Create(MapCacheFile)
	if err != nil {
		fmt.Printf("Could not create cache file: %v\n", err)
		return
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(s.maps); err != nil {
		fmt.Printf("Error encoding maps cache: %v\n", err)
	} else {
		fmt.Println("Maps cache saved successfully.")
	}
}

func (s *MapService) LoadCache() bool {
	if _, err := os.Stat(MapCacheFile); os.IsNotExist(err) {
		return false
	}

	file, err := os.Open(MapCacheFile)
	if err != nil {
		return false
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&s.maps); err != nil {
		fmt.Printf("Error decoding maps cache: %v\n", err)
		s.maps = make(map[int]*model.Map)
		return false
	}

	// Re-initialize non-exported or non-cached fields
	for _, m := range s.maps {
		m.Characters = make(map[int16]*model.Character)
		m.Npcs = make(map[int16]*model.WorldNPC)
	}

	return true
}

func (s *MapService) GetLoadedMaps() []int {
	ids := make([]int, 0, len(s.maps))
	for id := range s.maps {
		ids = append(ids, id)
	}
	return ids
}

func (s *MapService) LoadMap(id int) error {
	m, err := s.mapDAO.LoadMap(id)
	if err != nil {
		return err
	}
	m.Characters = make(map[int16]*model.Character)
	m.Npcs = make(map[int16]*model.WorldNPC)
	s.resolveMapEntities(m)
	s.maps[m.Id] = m
	return nil
}

func (s *MapService) UnloadMap(id int) {
	m := s.GetMap(id)
	if m == nil {
		return
	}

	// Remove all NPCs associated with this map
	m.RLock()
	var npcsToRemove []*model.WorldNPC
	for _, npc := range m.Npcs {
		npcsToRemove = append(npcsToRemove, npc)
	}
	m.RUnlock()

	for _, npc := range npcsToRemove {
		npc.Respawn = false
		s.npcService.RemoveNPC(npc, s)
	}

	delete(s.maps, id)
}

func (s *MapService) ReloadMap(id int) error {
	s.UnloadMap(id)
	return s.LoadMap(id)
}

func (s *MapService) resolveMapEntities(m *model.Map) {
	objectsFound := 0
	npcsFound := 0
	// Resolve objects and NPCs from map file
	for i := range m.Tiles {
		tile := &m.Tiles[i]
		x := i % model.MapWidth
		y := i / model.MapWidth

		if tile.ObjectID > 0 {
			obj := s.objectService.GetObject(tile.ObjectID)
			if obj != nil {
				tile.Object = &model.WorldObject{
					Object: obj,
					Amount: tile.ObjectAmount,
				}
				objectsFound++

				// Ensure door blocking is synchronized with object state
				if obj.Type == model.OTDoor {
					isClosed := obj.OpenIndex != 0
					tile.Blocked = isClosed
					if x > 0 {
						m.GetTile(x-1, y).Blocked = isClosed
					}
				}
			} else {
				fmt.Printf("Map %d: Could not resolve object ID %d at tile %d\n", m.Id, tile.ObjectID, i)
			}
		}

		if tile.NPCID > 0 {
			pos := model.Position{X: byte(x), Y: byte(y), Map: m.Id}
			worldNpc := s.npcService.SpawnNpc(tile.NPCID, pos)
			if worldNpc != nil {
				tile.NPC = worldNpc
				m.Npcs[worldNpc.Index] = worldNpc
				npcsFound++
			} else {
				fmt.Printf("Map %d: Could not resolve NPC ID %d at tile %d\n", m.Id, tile.NPCID, i)
			}
		}
	}

	if objectsFound > 0 || npcsFound > 0 {
		// fmt.Printf("Map %d: Resolved %d objects and %d NPCs on ground.\n", m.Id, objectsFound, npcsFound)
	}
}

func (s *MapService) GetMap(id int) *model.Map {
	return s.maps[id]
}

func (s *MapService) PutCharacterAtPos(char *model.Character, pos model.Position) {
	// 1. Remove from wherever it was before
	s.RemoveCharacter(char)

	// 2. Add to new position
	m := s.GetMap(pos.Map)
	if m == nil {
		return
	}

	m.Lock()
	defer m.Unlock()

	tile := m.GetTile(int(pos.X), int(pos.Y))
	if tile.Character != nil && tile.Character != char {
		fmt.Printf("Warning: PutCharacterAtPos overwriting character at %d,%d\n", pos.X, pos.Y)
	}
	if tile.NPC != nil {
		fmt.Printf("Warning: PutCharacterAtPos overwriting NPC at %d,%d\n", pos.X, pos.Y)
		tile.NPC = nil // NPCs are removed if a character teleports on top of them
	}

	m.Characters[char.CharIndex] = char
	tile.Character = char
	char.Position = pos
}

func (s *MapService) RemoveCharacter(char *model.Character) {
	m := s.GetMap(char.Position.Map)
	if m != nil {
		m.Lock()
		delete(m.Characters, char.CharIndex)
		tile := m.GetTile(int(char.Position.X), int(char.Position.Y))
		if tile.Character == char {
			tile.Character = nil
		}
		m.Unlock()
	}
}

func (s *MapService) ForEachCharacter(mapID int, f func(*model.Character)) {
	m := s.GetMap(mapID)
	if m == nil {
		return
	}
	m.RLock()
	defer m.RUnlock()
	for _, char := range m.Characters {
		f(char)
	}
}

func (s *MapService) ForEachNpc(mapID int, f func(*model.WorldNPC)) {
	m := s.GetMap(mapID)
	if m == nil {
		return
	}
	m.RLock()
	defer m.RUnlock()
	for _, npc := range m.Npcs {
		f(npc)
	}
}

func (s *MapService) GetObjectAt(pos model.Position) *model.WorldObject {
	m := s.GetMap(pos.Map)
	if m == nil {
		return nil
	}
	return m.GetTile(int(pos.X), int(pos.Y)).Object
}

func (s *MapService) PutObject(pos model.Position, obj *model.WorldObject) {
	m := s.GetMap(pos.Map)
	if m == nil {
		return
	}
	m.GetTile(int(pos.X), int(pos.Y)).Object = obj
}

func (s *MapService) RemoveObject(pos model.Position) {
	m := s.GetMap(pos.Map)
	if m == nil {
		return
	}
	m.GetTile(int(pos.X), int(pos.Y)).Object = nil
}

func (s *MapService) GetNPCAt(pos model.Position) *model.WorldNPC {
	m := s.GetMap(pos.Map)
	if m == nil {
		return nil
	}
	return m.GetTile(int(pos.X), int(pos.Y)).NPC
}

func (s *MapService) RemoveNPC(npc *model.WorldNPC) {
	m := s.GetMap(npc.Position.Map)
	if m != nil {
		m.Lock()
		delete(m.Npcs, npc.Index)
		tile := m.GetTile(int(npc.Position.X), int(npc.Position.Y))
		if tile.NPC == npc {
			tile.NPC = nil
		}
		m.Unlock()
	}
}

func (s *MapService) IsInPlayableArea(x, y int) bool {
	return x >= 5 && x <= 95 && y >= 5 && y <= 95
}

func (s *MapService) MoveNpc(npc *model.WorldNPC, newPos model.Position) bool {
	// Boundary checks
	if !s.IsInPlayableArea(int(newPos.X), int(newPos.Y)) {
		return false
	}

	oldPos := npc.Position
	mOld := s.GetMap(oldPos.Map)
	mNew := s.GetMap(newPos.Map)

	if mOld == nil || mNew == nil {
		return false
	}

	// Same map movement
	if mOld == mNew {
		mOld.Lock()
		defer mOld.Unlock()

		targetTile := mOld.GetTile(int(newPos.X), int(newPos.Y))
		if targetTile.Blocked || targetTile.NPC != nil || targetTile.Character != nil {
			return false
		}

		// Water check: NPCs cannot walk on water unless there is a bridge
		hasBridge := targetTile.Layer2 > 0
		if targetTile.IsWater && !hasBridge {
			return false
		}

		// Remove from old tile
		oldTile := mOld.GetTile(int(oldPos.X), int(oldPos.Y))
		if oldTile.NPC == npc {
			oldTile.NPC = nil
		}

		// Add to new tile
		targetTile.NPC = npc
		return true
	}

	// Cross-map movement (Rare for NPCs but possible)
	// To avoid deadlock, always lock in order of Map ID
	first, second := mOld, mNew
	if mOld.Id > mNew.Id {
		first, second = mNew, mOld
	}

	first.Lock()
	second.Lock()
	defer first.Unlock()
	defer second.Unlock()

	targetTile := mNew.GetTile(int(newPos.X), int(newPos.Y))
	if targetTile.Blocked || targetTile.NPC != nil || targetTile.Character != nil {
		return false
	}

	// Water check: NPCs cannot walk on water unless there is a bridge
	hasBridge := targetTile.Layer2 > 0
	if targetTile.IsWater && !hasBridge {
		return false
	}

	// Remove from old map
	oldTile := mOld.GetTile(int(oldPos.X), int(oldPos.Y))
	if oldTile.NPC == npc {
		oldTile.NPC = nil
	}
	delete(mOld.Npcs, npc.Index)

	// Add to new map
	mNew.Npcs[npc.Index] = npc
	targetTile.NPC = npc

	return true
}

func (s *MapService) MoveCharacterTo(char *model.Character, heading model.Heading) (model.Position, bool) {
	oldPos := char.Position
	newPos := oldPos
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
	if !s.IsInPlayableArea(int(newPos.X), int(newPos.Y)) {
		return char.Position, false
	}

	// Check if tile is blocked
	gameMap := s.GetMap(newPos.Map)
	if gameMap == nil {
		return char.Position, false
	}

	gameMap.Lock()
	defer gameMap.Unlock()

	tile := gameMap.GetTile(int(newPos.X), int(newPos.Y))

	// Map static blocking
	if tile.Blocked {
		return char.Position, false
	}

	// Sailing Logic
	hasBridge := tile.Layer2 > 0 || tile.Layer3 > 0
	if char.Sailing {
		if !tile.IsWater || hasBridge {
			return char.Position, false
		}
	} else {
		if tile.IsWater && !hasBridge {
			return char.Position, false
		}
	}

	// Occupancy check
	if tile.Character != nil || tile.NPC != nil {
		return char.Position, false
	}

	// Perform the move on the map
	// 1. Remove from old tile
	oldTile := gameMap.GetTile(int(oldPos.X), int(oldPos.Y))
	if oldTile.Character == char {
		oldTile.Character = nil
	}

	// 2. Add to new tile
	tile.Character = char

	char.Heading = heading
	char.Position = newPos

	// Ensure it's in the map's characters list (should already be there if same map)
	gameMap.Characters[char.CharIndex] = char

	return newPos, true
}

func (s *MapService) IsSafeZone(pos model.Position) bool {
	m := s.GetMap(pos.Map)
	if m == nil {
		return true // Assume safe if map not found? Or false? Usually false for safety.
	}
	tile := m.GetTile(int(pos.X), int(pos.Y))
	return tile.Trigger == model.TriggerSafeZone
}

func (s *MapService) IsPkMap(mapID int) bool {
	m := s.GetMap(mapID)
	if m == nil {
		return false
	}
	return m.Pk
}

func (s *MapService) IsInvalidPosition(pos model.Position) bool {
	m := s.GetMap(pos.Map)
	if m == nil {
		return true
	}
	tile := m.GetTile(int(pos.X), int(pos.Y))
	return tile.Trigger == model.TriggerInvalidPosition
}

func (s *MapService) IsTileEmpty(mapID int, x, y int) bool {
	m := s.GetMap(mapID)
	if m == nil {
		return false
	}
	if !s.IsInPlayableArea(x, y) {
		return false
	}
	tile := m.GetTile(x, y)

	// Water check: NPCs cannot be on water unless there is a bridge
	hasBridge := tile.Layer2 > 0
	if tile.IsWater && !hasBridge {
		return false
	}

	return !tile.Blocked && tile.Character == nil && tile.NPC == nil
}

func (s *MapService) SpawnNpcInMap(npcID int, mapID int) *model.WorldNPC {
	m := s.GetMap(mapID)
	if m == nil {
		return nil
	}

	// Try to find a random empty tile
	for i := 0; i < 100; i++ {
		x := rand.Intn(model.MapWidth)
		y := rand.Intn(model.MapHeight)

		if !s.IsInPlayableArea(x, y) {
			continue
		}

		//Valid position check
		if s.IsInvalidPosition(
			model.Position{
				X:   byte(x),
				Y:   byte(y),
				Map: mapID,
			}) {
			continue
		}

		if s.IsTileEmpty(mapID, x, y) {
			pos := model.Position{X: byte(x), Y: byte(y), Map: mapID}
			worldNpc := s.npcService.SpawnNpc(npcID, pos)
			if worldNpc != nil {
				m.Lock()
				m.Npcs[worldNpc.Index] = worldNpc
				m.Unlock()
				m.GetTile(x, y).NPC = worldNpc
				return worldNpc
			}
		}
	}
	return nil
}
