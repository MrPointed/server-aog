package model

import "math"

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
	TileExit  *Position
	Character *Character
	Object    *WorldObject

	// Fields for initial loading from map files
	ObjectID     int
	ObjectAmount int
}

const (
	MapWidth  = 100
	MapHeight = 100
)

type Map struct {
	Id      int
	Name    string
	Version int16
	Tiles   []Tile
	Characters map[int16]*Character
}

func (m *Map) GetTile(x, y int) *Tile {
	return &m.Tiles[y*MapWidth+x]
}
