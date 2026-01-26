package incoming

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/service"
)

type GMCommandsPacket struct {
	GMService service.GmService
}

func (p *GMCommandsPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	commandID, err := buffer.Get()
	if err != nil {
		return false, nil
	}
    
    return p.GMService.HandleCommand(connection, commandID, buffer)
}
