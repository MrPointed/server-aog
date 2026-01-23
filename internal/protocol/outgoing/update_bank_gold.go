package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type UpdateBankGoldPacket struct {
	Gold int
}

func (p *UpdateBankGoldPacket) Write(buffer *network.DataBuffer) error {
	buffer.PutInt(int32(p.Gold))
	return nil
}
