package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type UpdateHungerAndThirstPacket struct {
	MinHunger int
	MaxHunger int
	MinThirst int
	MaxThirst int
}

func (p *UpdateHungerAndThirstPacket) Write(buffer *network.DataBuffer) error {
	buffer.Put(byte(p.MaxThirst))
	buffer.Put(byte(p.MinThirst))
	buffer.Put(byte(p.MaxHunger))
	buffer.Put(byte(p.MinHunger))
	return nil
}
