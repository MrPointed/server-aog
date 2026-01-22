package incoming

import (
	"fmt"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/service"
)

type TalkPacket struct {
	MessageService *service.MessageService
}

func (p *TalkPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
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

	fmt.Printf("TALK packet received from [%s]: %s\n", char.Name, message)

	// Broadcast to area
	p.MessageService.SendToArea(&outgoing.ChatOverHeadPacket{
		Message:   message,
		CharIndex: char.CharIndex,
		R:         255, G: 255, B: 255, // White
	}, char.Position)

	return true, nil
}

