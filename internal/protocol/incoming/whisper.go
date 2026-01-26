package incoming

import (
	"fmt"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/service"
)

type WhisperPacket struct {
	UserService    service.UserService
}

func (p *WhisperPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	if buffer.ReadableBytes() < 4 {
		return false, nil
	}

	targetIndex, _ := buffer.GetShort()
	message, err := buffer.GetUTF8String()
	if err != nil {
		return false, err
	}

	char := connection.GetUser()
	if char == nil {
		return true, nil
	}

	targetChar := p.UserService.GetCharacterByIndex(targetIndex)
	if targetChar != nil {
		targetConn := p.UserService.GetConnection(targetChar)
		if targetConn != nil {
			targetConn.Send(&outgoing.ConsoleMessagePacket{
				Message: fmt.Sprintf("%s> %s", char.Name, message),
				Font:    outgoing.TALK,
			})
		}
	} else {
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "El usuario no está conectado.",
			Font:    outgoing.INFO,
		})
	}

	return true, nil
}
