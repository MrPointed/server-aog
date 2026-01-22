package incoming

import (
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/service"
)

type ResurrectPacket struct {
	MapService     *service.MapService
	AreaService    *service.AreaService
	MessageService *service.MessageService
	BodyService    *service.CharacterBodyService
}

func (p *ResurrectPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	char := connection.GetUser()
	if char == nil || !char.Dead {
		return true, nil
	}

	// Look for Healer NPC nearby
	healerFound := false
	gameMap := p.MapService.GetMap(char.Position.Map)
	if gameMap != nil {
		// Let's iterate nearby tiles
		for dx := -3; dx <= 3; dx++ {
			for dy := -3; dy <= 3; dy++ {
				tx, ty := int(char.Position.X)+dx, int(char.Position.Y)+dy
				if tx < 0 || tx >= 100 || ty < 0 || ty >= 100 {
					continue
				}
				tile := gameMap.GetTile(tx, ty)
				if tile.NPC != nil && tile.NPC.NPC.Type == model.NTHealer {
					healerFound = true
					break
				}
			}
			if healerFound {
				break
			}
		}
	}

	if !healerFound {
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "No hay nadie aquí que pueda resucitarte.",
			Font:    outgoing.INFO,
		})
		return true, nil
	}

	// Resurrect!
	char.Dead = false
	char.Hp = char.MaxHp
	char.Head = char.OriginalHead
	char.Body = p.BodyService.GetBody(char.Race, char.Gender)

	connection.Send(&outgoing.ConsoleMessagePacket{
		Message: "Has sido resucitado.",
		Font:    outgoing.INFO,
	})

	// Sync self
	connection.Send(outgoing.NewUpdateUserStatsPacket(char))
	
	// Broadcast change
	p.MessageService.SendToArea(&outgoing.CharacterChangePacket{Character: char}, char.Position)

	return true, nil
}
