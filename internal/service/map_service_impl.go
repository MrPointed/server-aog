package service

import (
	"encoding/gob"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"sync"

	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/persistence"
)

const MapCacheFile = "resources/maps.cache"

type MapServiceImpl struct {
	mapDAO        persistence.MapRepository
	objectService ObjectService
	npcService    NpcService
	maps          map[int]*model.Map
}

func NewMapServiceImpl(mapDAO persistence.MapRepository, objectService ObjectService, npcService NpcService) MapService {
	// Register types for gob
	gob.Register(model.Position{})
	return &MapServiceImpl{
		mapDAO:        mapDAO,
		objectService: objectService,
		npcService:    npcService,
		maps:          make(map[int]*model.Map),
	}
}

func (s *MapServiceImpl) LoadMaps() {
	if s.LoadCache() {
		// We still need to resolve entities because pointers to Object/NPC definitions
		// are not cached and depend on the current objects.dat/npcs.dat
		for _, m := range s.maps {
			s.resolveMapEntities(m)
		}

		slog.Info(fmt.Sprintf("Successfully loaded %d maps.", len(s.maps)))
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
				slog.Error("Error loading map", "map_id", id, "error", err)
				return
			}
			m.InitEntities()

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

	slog.Info("Loaded maps", "count", len(s.maps))
}

func (s *MapServiceImpl) SaveCache() {
	file, err := os.Create(MapCacheFile)
	if err != nil {
		slog.Error("Could not create cache file", "error", err)
		return
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(s.maps); err != nil {
		slog.Error("Error encoding maps cache", "error", err)
	} else {
		slog.Info("Maps cache saved successfully.")
	}
}

func (s *MapServiceImpl) LoadCache() bool {
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
		slog.Error("Error decoding maps cache", "error", err)
		s.maps = make(map[int]*model.Map)
		return false
	}

	// Re-initialize non-exported or non-cached fields and clear entity pointers
	for _, m := range s.maps {
		m.InitEntities()

		for i := range m.Tiles {
			m.Tiles[i].NPC = nil
			m.Tiles[i].Character = nil
			m.Tiles[i].Object = nil
		}
	}

	return true
}

func (s *MapServiceImpl) GetLoadedMaps() []int {
	ids := make([]int, 0, len(s.maps))
	for id := range s.maps {
		ids = append(ids, id)
	}
	return ids
}

func (s *MapServiceImpl) LoadMap(id int) error {
	m, err := s.mapDAO.LoadMap(id)
	if err != nil {
		return err
	}
	m.InitEntities()

	s.resolveMapEntities(m)
	s.maps[m.Id] = m
	return nil
}

func (s *MapServiceImpl) UnloadMap(id int) {
	m := s.GetMap(id)
	if m == nil {
		return
	}

	// Remove all NPCs associated with this map
	var npcsToRemove []*model.WorldNPC
	m.View(func(m *model.Map) {
		for _, npc := range m.GetNpcs() {
			npcsToRemove = append(npcsToRemove, npc)
		}
	})

	for _, npc := range npcsToRemove {
		npc.Respawn = false
		s.npcService.RemoveNPC(npc, s)
	}

	delete(s.maps, id)
}

func (s *MapServiceImpl) ReloadMap(id int) error {
	s.UnloadMap(id)
	return s.LoadMap(id)
}

func (s *MapServiceImpl) resolveMapEntities(m *model.Map) {
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
				slog.Warn("Could not resolve object", "map_id", m.Id, "obj_id", tile.ObjectID, "tile", i)
			}
		}

		if tile.NPCID > 0 {
			pos := model.Position{X: byte(x), Y: byte(y), Map: m.Id}
			worldNpc := s.npcService.SpawnNpc(tile.NPCID, pos)
			if worldNpc != nil {
				tile.NPC = worldNpc
				m.AddNpc(worldNpc)
				npcsFound++
			} else {
				slog.Warn("Could not resolve NPC", "map_id", m.Id, "npc_id", tile.NPCID, "tile", i)
			}
		}
	}

	if objectsFound > 0 || npcsFound > 0 {
		// slog.Debug("Resolved entities on ground", "map_id", m.Id, "objects", objectsFound, "npcs", npcsFound)
	}
}

func (s *MapServiceImpl) GetMap(id int) *model.Map {
	return s.maps[id]
}

func (s *MapServiceImpl) PutCharacterAtPos(char *model.Character, pos model.Position) {
	// 1. Remove from wherever it was before
	s.RemoveCharacter(char)

	// 2. Add to new position
	m := s.GetMap(pos.Map)
	if m == nil {
		return
	}

	m.Modify(func(m *model.Map) {
		tile := m.GetTile(int(pos.X), int(pos.Y))
		if tile.Character != nil && tile.Character != char {
			slog.Warn("PutCharacterAtPos overwriting character", "x", pos.X, "y", pos.Y)
		}
		if tile.NPC != nil {
			slog.Warn("PutCharacterAtPos overwriting NPC", "x", pos.X, "y", pos.Y)
			tile.NPC = nil // NPCs are removed if a character teleports on top of them
		}

		m.AddCharacter(char)
		tile.Character = char
		char.Position = pos
	})
}

func (s *MapServiceImpl) RemoveCharacter(char *model.Character) {
	m := s.GetMap(char.Position.Map)
	if m != nil {
		m.Modify(func(m *model.Map) {
			m.RemoveCharacter(char.CharIndex)
			tile := m.GetTile(int(char.Position.X), int(char.Position.Y))
			if tile.Character == char {
				tile.Character = nil
			}
		})
	}
}

func (s *MapServiceImpl) ForEachCharacter(mapID int, f func(*model.Character)) {
	m := s.GetMap(mapID)
	if m == nil {
		return
	}
	m.View(func(m *model.Map) {
		for _, char := range m.GetCharacters() {
			f(char)
		}
	})
}

func (s *MapServiceImpl) ForEachNpc(mapID int, f func(*model.WorldNPC)) {
	m := s.GetMap(mapID)
	if m == nil {
		return
	}
	m.View(func(m *model.Map) {
		for _, npc := range m.GetNpcs() {
			f(npc)
		}
	})
}

func (s *MapServiceImpl) GetObjectAt(pos model.Position) *model.WorldObject {
	m := s.GetMap(pos.Map)
	if m == nil {
		return nil
	}
	return m.GetTile(int(pos.X), int(pos.Y)).Object
}

func (s *MapServiceImpl) PutObject(pos model.Position, obj *model.WorldObject) {
	m := s.GetMap(pos.Map)
	if m == nil {
		return
	}
	m.GetTile(int(pos.X), int(pos.Y)).Object = obj
}

func (s *MapServiceImpl) RemoveObject(pos model.Position) {
	m := s.GetMap(pos.Map)
	if m == nil {
		return
	}
	m.GetTile(int(pos.X), int(pos.Y)).Object = nil
}

func (s *MapServiceImpl) GetNPCAt(pos model.Position) *model.WorldNPC {
	m := s.GetMap(pos.Map)
	if m == nil {
		return nil
	}
	return m.GetTile(int(pos.X), int(pos.Y)).NPC
}

func (s *MapServiceImpl) RemoveNPC(npc *model.WorldNPC) {
	m := s.GetMap(npc.Position.Map)
	if m != nil {
		m.Modify(func(m *model.Map) {
			m.RemoveNpc(npc.Index)
			tile := m.GetTile(int(npc.Position.X), int(npc.Position.Y))
			if tile.NPC == npc {
				tile.NPC = nil
			}
		})
	}
}

func (s *MapServiceImpl) IsInPlayableArea(x, y int) bool {
	return x >= 5 && x <= 95 && y >= 5 && y <= 95
}

func (s *MapServiceImpl) MoveNpc(npc *model.WorldNPC, newPos model.Position) bool {
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

	var success bool
	model.ModifyTwo(mOld, mNew, func(mOld, mNew *model.Map) {
		targetTile := mNew.GetTile(int(newPos.X), int(newPos.Y))
		if targetTile.Blocked || targetTile.NPC != nil || targetTile.Character != nil {
			return
		}

		// Water check: NPCs cannot walk on water unless there is a bridge
		hasBridge := targetTile.Layer2 > 0
		if targetTile.IsWater && !hasBridge {
			return
		}

		// Remove from old map
		oldTile := mOld.GetTile(int(oldPos.X), int(oldPos.Y))
		if oldTile.NPC == npc {
			oldTile.NPC = nil
		}

		if mOld != mNew {
			mOld.RemoveNpc(npc.Index)
			mNew.AddNpc(npc)
		}

		// Add to new map/tile
		targetTile.NPC = npc
		success = true
	})

	return success
}

func (s *MapServiceImpl) MoveCharacterTo(char *model.Character, heading model.Heading) (model.Position, bool) {
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

	if char.Paralyzed || char.Immobilized {
		return char.Position, false
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

	gameMap.Modify(func(gameMap *model.Map) {
		tile := gameMap.GetTile(int(newPos.X), int(newPos.Y))

		// Map static blocking
		if tile.Blocked {
			return
		}

		// Sailing Logic
		hasBridge := tile.Layer2 > 0 || tile.Layer3 > 0
		if char.Sailing {
			if !tile.IsWater || hasBridge {
				return
			}
		} else {
			if tile.IsWater && !hasBridge {
				return
			}
		}

		// Occupancy check
		if tile.Character != nil || tile.NPC != nil {
			return
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
		gameMap.AddCharacter(char)

		newPos = char.Position // Update newPos just in case, though it was local
	})

	// If position changed, it means success (since newPos started as oldPos modification but we return based on char.Position change? No, better use a flag or check if char.Position changed)
	if char.Position == newPos { // Wait, newPos was derived from oldPos at start of function
		return newPos, true
	}
	return char.Position, false
}

func (s *MapServiceImpl) IsSafeZone(pos model.Position) bool {
	m := s.GetMap(pos.Map)
	if m == nil {
		return true // Assume safe if map not found? Or false? Usually false for safety.
	}
	tile := m.GetTile(int(pos.X), int(pos.Y))
	return tile.Trigger == model.TriggerSafeZone
}

func (s *MapServiceImpl) IsPkMap(mapID int) bool {
	m := s.GetMap(mapID)
	if m == nil {
		return false
	}
	return m.Pk
}

func (s *MapServiceImpl) IsInvalidPosition(pos model.Position) bool {
	m := s.GetMap(pos.Map)
	if m == nil {
		return true
	}
	tile := m.GetTile(int(pos.X), int(pos.Y))
	return tile.Trigger == model.TriggerInvalidPosition
}

func (s *MapServiceImpl) IsTileEmpty(mapID int, x, y int) bool {
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

func (s *MapServiceImpl) IsBlocked(mapID, x, y int) bool {
	m := s.GetMap(mapID)
	if m == nil {
		return true
	}
	if !s.IsInPlayableArea(x, y) {
		return true
	}
	tile := m.GetTile(x, y)
	return tile.Blocked
}

func (s *MapServiceImpl) SpawnNpcInMap(npcID int, mapID int) *model.WorldNPC {
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
				m.Modify(func(m *model.Map) {
					m.AddNpc(worldNpc)
					m.GetTile(x, y).NPC = worldNpc
				})
				return worldNpc
			}
		}
	}
	return nil
}
