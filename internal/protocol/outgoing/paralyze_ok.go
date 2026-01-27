package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type ParalyzeOkPacket struct {
}

func (p *ParalyzeOkPacket) Write(buffer *network.DataBuffer) error {
	return nil
}
