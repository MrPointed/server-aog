package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type BankInitPacket struct {
	Gold int
}

func (p *BankInitPacket) Write(buffer *network.DataBuffer) error {
	buffer.PutInt(int32(p.Gold))
	return nil
}
