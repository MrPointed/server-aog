package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type NavigateTogglePacket struct{}

func (p *NavigateTogglePacket) Write(buf *network.DataBuffer) error {
	// No data body
	return nil
}
