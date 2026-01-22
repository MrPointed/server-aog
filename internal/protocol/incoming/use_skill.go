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
	// Standard CP_UseSkill usually sends the Skill ID (Byte)
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
			// These require a target
			// Note: Magic is usually via CastSpell, but if requested here, we ask for target?
			// Standard AO doesn't use UseSkill for Magic, but let's allow targeting to debug user issue.
			connection.Send(&outgoing.WorkRequestTargetPacket{Skill: skillID})
		
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

