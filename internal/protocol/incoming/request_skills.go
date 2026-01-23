package incoming

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type RequestSkillsPacket struct {
}

func (p *RequestSkillsPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	char := connection.GetUser()
	if char == nil {
		return true, nil
	}

	connection.Send(&outgoing.SendSkillsPacket{
		Archetype: char.Archetype,
		Skills:    char.Skills,
	})

	return true, nil
}
