package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type FamePacket struct {
	Assassin int32
	Bandit   int32
	Burgher  int32
	Thief    int32
	Noble    int32
	Plebeian int32
	Average  int32
}

func (p *FamePacket) Write(buffer *network.DataBuffer) error {
	buffer.Put(0) // Dummy byte
	buffer.PutInt(p.Assassin)
	buffer.PutInt(p.Bandit)
	buffer.PutInt(p.Burgher)
	buffer.PutInt(p.Thief)
	buffer.PutInt(p.Noble)
	buffer.PutInt(p.Plebeian)
	buffer.PutInt(p.Average)
	return nil
}
