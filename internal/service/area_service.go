package service

import (
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type AreaService struct {
	mapService  *MapService
	userService *UserService
}

func NewAreaService(mapService *MapService, userService *UserService) *AreaService {
	return &AreaService{
		mapService:  mapService,
		userService: userService,
	}
}

// In AO, areas are 9x9 tiles for the client, though server can use 18x18.
// We'll use 9 to stay in sync with the client's cleanup logic.
func (s *AreaService) GetArea(x, y byte) (int, int) {
	return int(x) / 9, int(y) / 9
}

func (s *AreaService) InRange(p1, p2 model.Position) bool {
	if p1.Map != p2.Map {
		return false
	}

	// Traditional AO logic: check if p2 is in any of the 9 areas around p1's area
	ax1, ay1 := s.GetArea(p1.X, p1.Y)
	ax2, ay2 := s.GetArea(p2.X, p2.Y)

	dx := ax1 - ax2
	dy := ay1 - ay2

	return dx >= -1 && dx <= 1 && dy >= -1 && dy <= 1
}

func (s *AreaService) BroadcastNearby(origin *model.Character, packet protocol.OutgoingPacket) {
	s.mapService.ForEachCharacter(origin.Position.Map, func(char *model.Character) {
		if char != origin {
			if s.InRange(origin.Position, char.Position) {
				conn := s.userService.GetConnection(char)
				if conn != nil {
					conn.Send(packet)
				}
			}
		}
	})
}

func (s *AreaService) BroadcastToArea(pos model.Position, packet protocol.OutgoingPacket) {
	s.mapService.ForEachCharacter(pos.Map, func(char *model.Character) {
		if s.InRange(pos, char.Position) {
			conn := s.userService.GetConnection(char)
			if conn != nil {
				conn.Send(packet)
			}
		}
	})
}

func (s *AreaService) NotifyNpcMovement(npc *model.WorldNPC, oldPos model.Position) {
	s.mapService.ForEachCharacter(npc.Position.Map, func(other *model.Character) {
		connOther := s.userService.GetConnection(other)
		if connOther == nil {
			return
		}

		wasInRange := s.InRange(oldPos, other.Position)
		isInRange := s.InRange(npc.Position, other.Position)

		if wasInRange && isInRange {
			// Just move
			connOther.Send(&outgoing.CharacterMovePacket{
				CharIndex: npc.Index,
				X:         npc.Position.X,
				Y:         npc.Position.Y,
				Heading:   npc.Heading,
			})
		} else if wasInRange && !isInRange {
			// NPC disappeared for this player
			connOther.Send(&outgoing.CharacterRemovePacket{CharIndex: npc.Index})
		} else if !wasInRange && isInRange {
			// NPC appeared for this player
			connOther.Send(&outgoing.NpcCreatePacket{Npc: npc})
		}
	})
}

func (s *AreaService) NotifyMovement(char *model.Character, oldPos model.Position) {
	s.mapService.ForEachCharacter(char.Position.Map, func(other *model.Character) {
		if other == char {
			return
		}

		connOther := s.userService.GetConnection(other)
		connMe := s.userService.GetConnection(char)

		wasInRange := s.InRange(oldPos, other.Position)
		isInRange := s.InRange(char.Position, other.Position)

		if wasInRange && isInRange {
			// Just move
			if connOther != nil {
				connOther.Send(&outgoing.CharacterMovePacket{
					CharIndex: char.CharIndex,
					X:         char.Position.X,
					Y:         char.Position.Y,
					Heading:   char.Heading,
				})
			}
		} else if wasInRange && !isInRange {
			// He was in range but now he's not.
			// I should see him disappear, and he should see me disappear.
			if connOther != nil {
				connOther.Send(&outgoing.CharacterRemovePacket{CharIndex: char.CharIndex})
			}
			if connMe != nil {
				connMe.Send(&outgoing.CharacterRemovePacket{CharIndex: other.CharIndex})
			}
		} else if !wasInRange && isInRange {
			// He wasn't in range but now he is.
			// I should see him appear, and he should see me appear.
			if connOther != nil {
				connOther.Send(&outgoing.CharacterCreatePacket{Character: char})
			}
			if connMe != nil {
				connMe.Send(&outgoing.CharacterCreatePacket{Character: other})
			}
		}
	})
}

func (s *AreaService) SendAreaState(char *model.Character) {
	conn := s.userService.GetConnection(char)
	if conn == nil {
		return
	}

	gameMap := s.mapService.GetMap(char.Position.Map)
	if gameMap == nil {
		return
	}

	gameMap.Mu.RLock()
	defer gameMap.Mu.RUnlock()

	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			tile := gameMap.GetTile(x, y)

			// Objects
			if tile.Object != nil {
				objPos := model.Position{X: byte(x), Y: byte(y), Map: char.Position.Map}
				if s.InRange(char.Position, objPos) {
					conn.Send(&outgoing.ObjectCreatePacket{
						X:            byte(x),
						Y:            byte(y),
						GraphicIndex: int16(tile.Object.Object.GraphicIndex),
					})

					// If it's a door, sync the blocking status as it might have changed from the static map file
					if tile.Object.Object.Type == model.OTDoor {
						conn.Send(&outgoing.BlockPositionPacket{
							X:       byte(x),
							Y:       byte(y),
							Blocked: tile.Blocked,
						})
						// Also sync the left tile which is part of the same door
						if x > 0 {
							leftTile := gameMap.GetTile(x-1, y)
							conn.Send(&outgoing.BlockPositionPacket{
								X:       byte(x - 1),
								Y:       byte(y),
								Blocked: leftTile.Blocked,
							})
						}
					}
				}
			}

			// NPCs
			if tile.NPC != nil {
				if s.InRange(char.Position, tile.NPC.Position) {
					conn.Send(&outgoing.NpcCreatePacket{Npc: tile.NPC})
				}
			}
		}
	}

	// Characters
	s.mapService.ForEachCharacter(char.Position.Map, func(other *model.Character) {
		if other != char {
			if s.InRange(char.Position, other.Position) {
				conn.Send(&outgoing.CharacterCreatePacket{Character: other})
			}
		}
	})
}
