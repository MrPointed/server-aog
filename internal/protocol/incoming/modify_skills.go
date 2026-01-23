package incoming

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type ModifySkillsPacket struct {
}

func (p *ModifySkillsPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	char := connection.GetUser()
	if char == nil {
		return true, nil
	}

	// Read 20 skills
	var pointsToSpend [20]int
	totalSpent := 0
	for i := 0; i < 20; i++ {
		val, err := buffer.Get()
		if err != nil {
			return true, err
		}
		pointsToSpend[i] = int(val)
		totalSpent += pointsToSpend[i]
	}

	if totalSpent > char.SkillPoints {
		return true, nil
	}

	if totalSpent <= 0 {
		return true, nil
	}

	for i := 0; i < 20; i++ {
		if pointsToSpend[i] > 0 {
			skillID := model.Skill(i + 1)
			char.Skills[skillID] += pointsToSpend[i]
			if char.Skills[skillID] > 100 {
				char.Skills[skillID] = 100
			}
		}
	}

	char.SkillPoints -= totalSpent
	
	// Sync back
	connection.Send(outgoing.NewUpdateUserStatsPacket(char))
	connection.Send(&outgoing.SendSkillsPacket{
		Archetype:   char.Archetype,
		Skills:      char.Skills,
		SkillPoints: char.SkillPoints,
	})

	return true, nil
}
