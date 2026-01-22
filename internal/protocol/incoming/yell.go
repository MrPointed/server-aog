package incoming

import (
	"fmt"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/service"
)

type YellPacket struct {
	MessageService *service.MessageService
}

func (p *YellPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	if buffer.ReadableBytes() < 2 {
		return false, nil
	}

	message, err := buffer.GetUTF8String()
	if err != nil {
		return false, err
	}

	char := connection.GetUser()
	if char == nil {
		return true, nil
	}

	fmt.Printf("YELL packet received from [%s]: %s\n", char.Name, message)

	// Broadcast to map
	p.MessageService.SendToMap(&outgoing.ChatOverHeadPacket{
		Message:   message,
		CharIndex: char.CharIndex,
		R:         255, G: 255, B: 255, // White
	}, char.Position.Map)

	return true, nil
}