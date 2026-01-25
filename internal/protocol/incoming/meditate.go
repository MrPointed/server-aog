package incoming

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/service"
)

type MeditatePacket struct {
	AreaService *service.AreaService
}

func (p *MeditatePacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	char := connection.GetUser()
	if char == nil || char.Dead {
		return true, nil
	}

	char.Meditating = !char.Meditating
	
	// Notify the user about meditation toggle (client will probably show FX)
	connection.Send(&outgoing.MeditateTogglePacket{})

	// Notify the area about the meditation toggle
	p.AreaService.BroadcastToArea(char.Position, &outgoing.MeditateTogglePacket{})

	if char.Meditating {
		connection.Send(&outgoing.ConsoleMessagePacket{Message: "Te concentras...", Font: outgoing.INFO})
		// Start meditation FX (4 is common for meditation in AO, -1 for infinite loops)
		fxPacket := &outgoing.CreateFxPacket{
			CharIndex: char.CharIndex,
			FxID:      4,
			Loops:     -1,
		}
		connection.Send(fxPacket)
		p.AreaService.BroadcastNearby(char, fxPacket)
	} else {
		connection.Send(&outgoing.ConsoleMessagePacket{Message: "Dejas de meditar.", Font: outgoing.INFO})
		// Stop meditation FX
		fxPacket := &outgoing.CreateFxPacket{
			CharIndex: char.CharIndex,
			FxID:      0,
			Loops:     0,
		}
		connection.Send(fxPacket)
		p.AreaService.BroadcastNearby(char, fxPacket)
	}
	
	return true, nil
}
