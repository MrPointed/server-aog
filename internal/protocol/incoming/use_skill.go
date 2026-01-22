package incoming

import (
	"fmt"

	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type UseSkillPacket struct {
}

func (p *UseSkillPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	if buffer.ReadableBytes() < 1 {
		return false, nil
	}

	skillIDByte, _ := buffer.Get()
	skillID := model.Skill(skillIDByte)

	user := connection.GetUser()
	if user != nil {
		fmt.Printf("User %s requested UseSkill for Skill ID: %d\n", user.Name, skillID)

		switch skillID {
		case model.Steal, model.Tame, model.Magic:
			connection.Send(&outgoing.SkillRequestTargetPacket{Skill: skillID})

		case model.Meditate:
			// Toggle meditate
			// TODO: Implement Meditate logic
			connection.Send(&outgoing.ConsoleMessagePacket{Message: "Te concentras...", Font: outgoing.INFO})

		case model.Hiding:
			// Toggle hiding
			// TODO: Implement Hiding logic
			connection.Send(&outgoing.ConsoleMessagePacket{Message: "Te ocultas...", Font: outgoing.INFO})

		default:
			// Others
			connection.Send(&outgoing.ConsoleMessagePacket{Message: fmt.Sprintf("Skill %d no implementada en UseSkill.", skillID), Font: outgoing.INFO})
		}
	}

	return true, nil
}
