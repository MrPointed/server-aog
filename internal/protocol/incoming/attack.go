package incoming

import (
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/service"
)

type AttackPacket struct {
	MapService    *service.MapService
	CombatService *service.CombatService
	AreaService   *service.AreaService
}

func (p *AttackPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	char := connection.GetUser()
	if char == nil {
		return true, nil
	}

	if char.Dead {
		return true, nil
	}

	if char.Meditating {
		char.Meditating = false
		connection.Send(&outgoing.MeditateTogglePacket{})
		p.AreaService.BroadcastToArea(char.Position, &outgoing.MeditateTogglePacket{})
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "Dejas de meditar.",
			Font:    outgoing.INFO,
		})
		// Stop meditation FX
		fxPacket := &outgoing.CreateFxPacket{
			CharIndex: char.CharIndex,
			FxID:      0,
			Loops:     0,
		}
		connection.Send(fxPacket)
		p.AreaService.BroadcastNearby(char, fxPacket)
	}

	// Find target in front
	targetPos := char.Position
	switch char.Heading {
	case model.North:
		targetPos.Y--
	case model.South:
		targetPos.Y++
	case model.East:
		targetPos.X++
	case model.West:
		targetPos.X--
	}

	// Check for character
	gameMap := p.MapService.GetMap(targetPos.Map)
	if gameMap == nil {
		return true, nil
	}

	tile := gameMap.GetTile(int(targetPos.X), int(targetPos.Y))
	if tile.Character != nil {
		p.CombatService.ResolveAttack(char, tile.Character)
	} else if tile.NPC != nil {
		p.CombatService.ResolveAttack(char, tile.NPC)
	}

	return true, nil
}
