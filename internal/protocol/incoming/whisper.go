package incoming

import (
	"fmt"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
)

type WhisperPacket struct{}

func (p *WhisperPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	if buffer.ReadableBytes() < 2 {
		return false, nil
	}

	// Assuming dedicated whisper packet might have target and message, 
	// but let's just read one string for now as in Java it's a placeholder.
	message, err := buffer.GetUTF8String()
	if err != nil {
		return false, err
	}

	char := connection.GetUser()
	name := "Unknown"
	if char != nil {
		name = char.Name
	}

	fmt.Printf("WHISPER packet received from [%s]: %s\n", name, message)

	return true, nil
}
