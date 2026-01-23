package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type UpdateGoldPacket struct {
	Gold int
}

func (p *UpdateGoldPacket) Write(buffer *network.DataBuffer) error {
	buffer.PutInt(int32(p.Gold))
	return nil
}
