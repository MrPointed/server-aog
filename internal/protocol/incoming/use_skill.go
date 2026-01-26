package incoming

import (
	"fmt"
	"time"

	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/service"
)

type UseSkillPacket struct {
	AreaService service.AreaService
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
			if !user.Dead {
						user.Meditating = !user.Meditating
						connection.Send(&outgoing.MeditateTogglePacket{})
						if user.Meditating {
				
					user.MeditatingSince = time.Now()
					user.LastMeditationRegen = time.Time{}
					connection.Send(&outgoing.ConsoleMessagePacket{Message: "Te concentras...", Font: outgoing.INFO})
				} else {
					connection.Send(&outgoing.ConsoleMessagePacket{Message: "Dejas de meditar.", Font: outgoing.INFO})
					fxPacket := &outgoing.CreateFxPacket{
						CharIndex: user.CharIndex,
						FxID:      0,
						Loops:     0,
					}
					connection.Send(fxPacket)
					p.AreaService.BroadcastNearby(user, fxPacket)
				}
			}

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
