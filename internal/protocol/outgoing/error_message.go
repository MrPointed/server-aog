package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type ErrorMessagePacket struct {
	Message string
}

func (p *ErrorMessagePacket) Write(buffer *network.DataBuffer) error {
	buffer.PutUTF8String(p.Message)
	return nil
}
