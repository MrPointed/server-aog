package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type BankingEndPacket struct {
}

func (p *BankingEndPacket) Write(buffer *network.DataBuffer) error {
	return nil
}
