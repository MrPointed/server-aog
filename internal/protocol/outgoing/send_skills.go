package outgoing

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/model"
)

type SendSkillsPacket struct {
	Archetype model.UserArchetype
	Skills    map[model.Skill]int
}

func (p *SendSkillsPacket) Write(buffer *network.DataBuffer) error {
	buffer.Put(byte(p.Archetype))

	// The client expects 20 skills (from the Skill enum)
	// Each skill is sent as: [Value (byte)][Percentage (byte)]
	for i := 1; i <= 20; i++ {
		val := p.Skills[model.Skill(i)]
		buffer.Put(byte(val))
		buffer.Put(0) // percentage placeholder
	}
	return nil
}
