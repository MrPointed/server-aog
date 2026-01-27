package incoming

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type RequestFamePacket struct {
}

func (p *RequestFamePacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	char := connection.GetUser()
	if char == nil {
		return true, nil
	}

	rep := char.Reputation

	// Calculate average (Promedio) used for criminal status
	// Status = (Noble + Plebe + Burgues) - (Asesino + Bandido + Ladron)
	average := int32((rep.Noble + rep.Commoner + rep.Burguer) - (rep.Assassin + rep.Bandit + rep.Thief))

	connection.Send(&outgoing.FamePacket{
		Assassin: int32(rep.Assassin),
		Bandit:   int32(rep.Bandit),
		Burgher:  int32(rep.Burguer),
		Thief:    int32(rep.Thief),
		Noble:    int32(rep.Noble),
		Plebeian: int32(rep.Commoner),
		Average:  average,
	})

	return true, nil
}
