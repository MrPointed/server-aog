package outgoing

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/model"
)

type AttributesPacket struct {
	Attributes map[model.Attribute]byte
}

func (p *AttributesPacket) Write(buffer *network.DataBuffer) error {
	buffer.Put(p.Attributes[model.Strength])
	buffer.Put(p.Attributes[model.Dexterity])
	buffer.Put(p.Attributes[model.Intelligence])
	buffer.Put(p.Attributes[model.Charisma])
	buffer.Put(p.Attributes[model.Constitution])
	return nil
}
