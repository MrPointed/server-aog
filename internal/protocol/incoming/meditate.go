package incoming

import (
	"time"

	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/service"
)

type MeditatePacket struct {
	AreaService service.AreaService
}

func (p *MeditatePacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	char := connection.GetUser()
	if char == nil || char.Dead {
		return true, nil
	}

	char.Meditating = !char.Meditating
	
	// Notify the user about meditation toggle
	connection.Send(&outgoing.MeditateTogglePacket{})

	if char.Meditating {
		char.MeditatingSince = time.Now()
		char.LastMeditationRegen = time.Time{} // Reset to indicate it hasn't "really" started
		connection.Send(&outgoing.ConsoleMessagePacket{Message: "Te concentras...", Font: outgoing.INFO})
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
