package incoming

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/model"
)

type RequestMiniStatsPacket struct {
}

func (p *RequestMiniStatsPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	char := connection.GetUser()
	if char == nil {
		return true, nil
	}

	// For now Role is determined by criminal status
	role := byte(0) // Citizen
	if char.Faccion.Criminal {
		role = 1 // Criminal
	}

	connection.Send(&outgoing.MiniStatsPacket{
		CriminalsKilled: int32(char.Kills[model.KillCriminals]),
		CitizensKilled:  int32(char.Kills[model.KillCitizens]),
		UsersKilled:     int32(char.Kills[model.KillUsers]),
		CreaturesKilled: int16(char.Kills[model.KillCreatures]),
		Role:            role,
		JailTime:        int32(char.JailTime),
	})

	return true, nil
}
