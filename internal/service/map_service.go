package service

import (
	"fmt"
	"math/rand"

	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/persistence"
)

type MapService struct {
	mapDAO        *persistence.MapDAO
	objectService *ObjectService
	npcService    *NpcService
	maps          map[int]*model.Map
}

func NewMapService(mapDAO *persistence.MapDAO, objectService *ObjectService, npcService *NpcService) *MapService {
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

		s.maps[m.Id] = m
		if objectsFound > 0 || npcsFound > 0 {
			fmt.Printf("Map %d: Resolved %d objects and %d NPCs on ground.\n", m.Id, objectsFound, npcsFound)
		}
	}
	fmt.Printf("Loaded %d maps\n", len(s.maps))
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
	
	m.Mu.Lock()
	defer m.Mu.Unlock()

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
		m.Mu.Lock()
		delete(m.Characters, char.CharIndex)
		tile := m.GetTile(int(char.Position.X), int(char.Position.Y))
		if tile.Character == char {
			tile.Character = nil
		}
		m.Mu.Unlock()
	}
}

func (s *MapService) ForEachCharacter(mapID int, f func(*model.Character)) {
	m := s.GetMap(mapID)
	if m == nil {
		return
	}
	m.Mu.RLock()
	defer m.Mu.RUnlock()
	for _, char := range m.Characters {
		f(char)
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
		m.Mu.Lock()
		delete(m.Npcs, npc.Index)
		tile := m.GetTile(int(npc.Position.X), int(npc.Position.Y))
		if tile.NPC == npc {
			tile.NPC = nil
		}
		m.Mu.Unlock()
	}
}

func (s *MapService) MoveNpc(npc *model.WorldNPC, newPos model.Position) bool {
	oldPos := npc.Position
	mOld := s.GetMap(oldPos.Map)
	mNew := s.GetMap(newPos.Map)

	if mOld == nil || mNew == nil {
		return false
	}

	// Same map movement
	if mOld == mNew {
		mOld.Mu.Lock()
		defer mOld.Mu.Unlock()

		targetTile := mOld.GetTile(int(newPos.X), int(newPos.Y))
		if targetTile.Blocked || targetTile.NPC != nil || targetTile.Character != nil {
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

	first.Mu.Lock()
	second.Mu.Lock()
	defer first.Mu.Unlock()
	defer second.Mu.Unlock()

	targetTile := mNew.GetTile(int(newPos.X), int(newPos.Y))
	if targetTile.Blocked || targetTile.NPC != nil || targetTile.Character != nil {
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

	// Boundary checks (simplified)
	if newPos.X < 0 || newPos.X >= 100 || newPos.Y < 0 || newPos.Y >= 100 {
		return char.Position, false
	}

	// Check if tile is blocked
	gameMap := s.GetMap(newPos.Map)
	if gameMap == nil {
		return char.Position, false
	}

	gameMap.Mu.Lock()
	defer gameMap.Mu.Unlock()

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
	if x <= 10 || x >= model.MapWidth-10 || y <= 10 || y >= model.MapHeight-10 {
		return false
	}
	tile := m.GetTile(x, y)
	return !tile.Blocked && tile.Character == nil && tile.NPC == nil
}

func (s *MapService) SpawnNpcInMap(npcID int, mapID int) *model.WorldNPC {
	m := s.GetMap(mapID)
	if m == nil {
		return nil
	}

	// Try to find a random empty tile
	for i := 11; i < 90; i++ {
		x := rand.Intn(model.MapWidth)
		y := rand.Intn(model.MapHeight)

		if s.IsTileEmpty(mapID, x, y) {
			pos := model.Position{X: byte(x), Y: byte(y), Map: mapID}
			worldNpc := s.npcService.SpawnNpc(npcID, pos)
			if worldNpc != nil {
				m.Mu.Lock()
				m.Npcs[worldNpc.Index] = worldNpc
				m.Mu.Unlock()
				m.GetTile(x, y).NPC = worldNpc
				return worldNpc
			}
		}
	}
	return nil
}
