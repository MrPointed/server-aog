package outgoing

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/model"
)

type AreaChangedPacket struct {
	Position model.Position
}

func (p *AreaChangedPacket) Write(buffer *network.DataBuffer) error {
	buffer.Put(p.Position.X + 1)
	buffer.Put(p.Position.Y + 1)
	return nil
}
