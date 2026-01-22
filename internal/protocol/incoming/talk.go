package incoming

import (
	"fmt"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
)

type TalkPacket struct{}

func (p *TalkPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	if buffer.ReadableBytes() < 2 {
		return false, nil
	}

	message, err := buffer.GetUTF8String() // Java getASCIIString uses readShort + new String
	if err != nil {
		return false, err
	}

	char := connection.GetUser()
	name := "Unknown"
	if char != nil {
		name = char.Name
	}

	fmt.Printf("TALK packet received from [%s]: %s\n", name, message)

	return true, nil
}

