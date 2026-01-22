package incoming

import (
	"fmt"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/service"
)

type LoginExistingCharacterPacket struct {
	LoginService *service.LoginService
}

func (p *LoginExistingCharacterPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	initialPos := buffer.Pos()
	
	username, err := buffer.GetUTF8String()
	if err != nil {
		return false, nil // Incomplete
	}

	password, err := buffer.GetUTF8String()
	if err != nil {
		buffer.SetPos(initialPos)
		return false, nil
	}

	if buffer.ReadableBytes() < 3 {
		buffer.SetPos(initialPos)
		return false, nil
	}

	v1, _ := buffer.Get()
	v2, _ := buffer.Get()
	v3, _ := buffer.Get()
	version := fmt.Sprintf("%d.%d.%d", v1, v2, v3)
	clientHash := "" // No client hash in this version

	err = p.LoginService.ConnectExistingCharacter(connection, username, password, version, clientHash)
	if err != nil {
		connection.Send(&outgoing.ErrorMessagePacket{Message: err.Error()})
		connection.Disconnect()
		return true, nil
	}

	return true, nil
}
