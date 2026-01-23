package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type MeditateTogglePacket struct {
}

func (p *MeditateTogglePacket) Write(buffer *network.DataBuffer) error {
	return nil
}
