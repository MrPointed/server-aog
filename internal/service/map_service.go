package service

import (
	"fmt"
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/persistence"
)

type MapService struct {
	mapDAO        *persistence.MapDAO
	objectService *ObjectService
	maps          map[int]*model.Map
}

func NewMapService(mapDAO *persistence.MapDAO, objectService *ObjectService) *MapService {
	return &MapService{
		mapDAO:        mapDAO,
		objectService: objectService,
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
		
		objectsFound := 0
		// Resolve objects from map file
		for i := range m.Tiles {
			tile := &m.Tiles[i]
			if tile.ObjectID > 0 {
				obj := s.objectService.GetObject(tile.ObjectID)
				if obj != nil {
					tile.Object = &model.WorldObject{
						Object: obj,
						Amount: tile.ObjectAmount,
					}
					objectsFound++
				} else {
					fmt.Printf("Map %d: Could not resolve object ID %d at tile %d\n", m.Id, tile.ObjectID, i)
				}
			}
		}
		
		s.maps[m.Id] = m
		if objectsFound > 0 {
			fmt.Printf("Map %d: Resolved %d default objects on ground.\n", m.Id, objectsFound)
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
	m.Characters[char.CharIndex] = char
	tile := m.GetTile(int(pos.X), int(pos.Y))
	tile.Character = char
	char.Position = pos
}

func (s *MapService) RemoveCharacter(char *model.Character) {
	m := s.GetMap(char.Position.Map)
	if m != nil {
		delete(m.Characters, char.CharIndex)
		tile := m.GetTile(int(char.Position.X), int(char.Position.Y))
		if tile.Character == char {
			tile.Character = nil
		}
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

func (s *MapService) MoveCharacterTo(char *model.Character, heading model.Heading) (model.Position, bool) {
	newPos := char.Position
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
	if gameMap != nil {
		tile := gameMap.GetTile(int(newPos.X), int(newPos.Y))
		if tile.Blocked || tile.Character != nil {
			return char.Position, false
		}
	}

	char.Heading = heading
	return newPos, true
}
