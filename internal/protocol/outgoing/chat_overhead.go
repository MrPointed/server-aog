package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type ChatOverHeadPacket struct {
	Message   string
	CharIndex int16
	R, G, B   byte
}

func (p *ChatOverHeadPacket) Write(buffer *network.DataBuffer) error {
	buffer.PutUTF8String(p.Message)
	buffer.PutShort(p.CharIndex)
	buffer.Put(p.R)
	buffer.Put(p.G)
	buffer.Put(p.B)
	return nil
}
