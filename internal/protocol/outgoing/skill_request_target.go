package outgoing

import (
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/network"
)

type SkillRequestTargetPacket struct {
	Skill model.Skill
}

func (p *SkillRequestTargetPacket) Write(buffer *network.DataBuffer) error {
	buffer.Put(byte(p.Skill))
	return nil
}
