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

func (m *Map) Lock() {
	m.mu.Lock()
}

func (m *Map) Unlock() {
	m.mu.Unlock()
}

func (m *Map) RLock() {
	m.mu.RLock()
}

func (m *Map) RUnlock() {
	m.mu.RUnlock()
}

func (m *Map) GetTile(x, y int) *Tile {
	return &m.Tiles[y*MapWidth+x]
}
