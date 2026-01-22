package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type LoggedPacket struct {
}

func (p *LoggedPacket) Write(buffer *network.DataBuffer) error {
	return nil
}
