package model

import (
	"math"
	"sync"
)

type Heading int

const (
	North Heading = iota
	East
	South
	West
)

type Trigger int

const (
	TriggerNone Trigger = iota
	TriggerUnderRoof
	Trigger2
	TriggerInvalidPosition
	TriggerSafeZone
	TriggerAntiPicket
	TriggerFightZone
)

type Position struct {
	X   byte
	Y   byte
	Map int
}

func (p Position) GetDistance(other Position) int {
	return int(math.Abs(float64(p.X)-float64(other.X)) + math.Abs(float64(p.Y)-float64(other.Y)))
}

type City struct {
	Map int
	X   byte
	Y   byte
}

type Tile struct {
	Trigger   Trigger
	Blocked   bool
	IsWater   bool
	IsLava    bool
	Layer2    int16
	Layer3    int16
	Layer4    int16
	TileExit  *Position
	Character *Character
	Object    *WorldObject
	NPC       *WorldNPC

	// Fields for initial loading from map files
	ObjectID     int
	ObjectAmount int
	NPCID        int
}

const (
	MapWidth  = 100
	MapHeight = 100
)

type Map struct {
	Id      int
	Name    string
	Version int16
	Pk      bool
	Tiles   []Tile
	Characters map[int16]*Character
	Npcs       map[int16]*WorldNPC
	mu         sync.RWMutex
}

// Modify executes the action under a write lock.
func (m *Map) Modify(action func(m *Map)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	action(m)
}

// View executes the action under a read lock.
func (m *Map) View(action func(m *Map)) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	action(m)
}

// ModifyTwo executes the action locking both maps safely (ordered by ID) to prevent deadlocks.
// If both maps are the same, it only locks once.
func ModifyTwo(m1, m2 *Map, action func(m1, m2 *Map)) {
	if m1 == m2 {
		m1.Modify(func(m *Map) {
			action(m, m)
		})
		return
	}

	first, second := m1, m2
	if m1.Id > m2.Id {
		first, second = m2, m1
	}

	first.mu.Lock()
	second.mu.Lock()
	defer first.mu.Unlock()
	defer second.mu.Unlock()

	action(m1, m2)
}

func (m *Map) GetTile(x, y int) *Tile {
	return &m.Tiles[y*MapWidth+x]
}
