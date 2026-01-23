package incoming

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type MeditatePacket struct {
}

func (p *MeditatePacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	char := connection.GetUser()
	if char == nil || char.Dead {
		return true, nil
	}

	char.Meditating = !char.Meditating
	
	// Notify the user about meditation toggle (client will probably show FX)
	connection.Send(&outgoing.MeditateTogglePacket{})
	
	return true, nil
}
