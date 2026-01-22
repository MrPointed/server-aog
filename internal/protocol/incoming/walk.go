package incoming

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/service"
)

type WalkPacket struct {
	MapService     *service.MapService
	MessageService *service.MessageService
	AreaService    *service.AreaService // Still needed for Area logic in Handle
}

func (p *WalkPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	if buffer.ReadableBytes() < 1 {
		return false, nil
	}

	rawHeading, _ := buffer.Get()
	heading := model.Heading(rawHeading - 1)

	char := connection.GetUser()
	if char == nil {
		return true, nil // Not logged in
	}

	if char.Paralyzed {
		return true, nil
	}

	if char.Meditating {
		char.Meditating = false
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "Dejas de meditar.",
			Font:    outgoing.INFO,
		})
	} else {
		oldPos := char.Position
		newPos, success := p.MapService.MoveCharacterTo(char, heading)
		
		if !success {
			// Send PosUpdate to sync client if move failed
			connection.Send(&outgoing.PosUpdatePacket{
				X: char.Position.X,
				Y: char.Position.Y,
			})
			return true, nil
		}

		// Update tile in MapService atomicity
		p.MapService.PutCharacterAtPos(char, newPos)

		// Check for Map Change (Tile Exit)
		gameMap := p.MapService.GetMap(newPos.Map)
		if gameMap != nil {
			tile := gameMap.GetTile(int(newPos.X), int(newPos.Y))
			if tile.TileExit != nil {
				targetMap := tile.TileExit.Map
				targetX := tile.TileExit.X
				targetY := tile.TileExit.Y

				// 1. Remove from current map
				p.MapService.RemoveCharacter(char)
				p.AreaService.BroadcastToArea(oldPos, &outgoing.CharacterRemovePacket{CharIndex: char.CharIndex})

				// 2. Set new position
				char.Position = model.Position{X: targetX, Y: targetY, Map: targetMap}

				// 3. Add to new map
				p.MapService.PutCharacterAtPos(char, char.Position)

				// 4. Send ChangeMap to client
				connection.Send(&outgoing.ChangeMapPacket{MapId: targetMap, Version: 0})

				// 5. Force Client Position Update
				connection.Send(&outgoing.PosUpdatePacket{
					X: char.Position.X,
					Y: char.Position.Y,
				})

				// 6. Send PlayMidi (TODO: Fetch correct MIDI)
				connection.Send(&outgoing.PlayWavePacket{Wave: 0}) // Placeholder

				// 7. Update Area
				connection.Send(&outgoing.AreaChangedPacket{Position: char.Position})
				p.AreaService.SendAreaState(char)
				p.AreaService.BroadcastToArea(char.Position, &outgoing.CharacterCreatePacket{Character: char})

				return true, nil
			}
		}

		// Check area change
		oldAX, oldAY := p.AreaService.GetArea(oldPos.X, oldPos.Y)
		newAX, newAY := p.AreaService.GetArea(char.Position.X, char.Position.Y)

		if oldAX != newAX || oldAY != newAY {
			connection.Send(&outgoing.AreaChangedPacket{Position: char.Position})
			
			// Sync objects in new range
			gameMap := p.MapService.GetMap(char.Position.Map)
			if gameMap != nil {
				for y := 0; y < 100; y++ {
					for x := 0; x < 100; x++ {
						tile := gameMap.GetTile(x, y)
						if tile.Object != nil {
							objPos := model.Position{X: byte(x), Y: byte(y), Map: char.Position.Map}
							if p.AreaService.InRange(char.Position, objPos) {
								connection.Send(&outgoing.ObjectCreatePacket{
									X:            byte(x),
									Y:            byte(y),
									GraphicIndex: int16(tile.Object.Object.GraphicIndex),
								})
							}
						}
						if tile.NPC != nil {
							if p.AreaService.InRange(char.Position, tile.NPC.Position) {
								connection.Send(&outgoing.NpcCreatePacket{Npc: tile.NPC})
							}
						}
					}
				}
			}
		}

		p.AreaService.NotifyMovement(char, oldPos)
		
		// Confirm to client
		connection.Send(&outgoing.PosUpdatePacket{
			X: char.Position.X,
			Y: char.Position.Y,
		})
	}

	return true, nil
}
