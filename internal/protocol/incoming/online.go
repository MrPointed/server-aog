package incoming

import (
	"fmt"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/service"
)

type OnlinePacket struct {
	UserService *service.UserService
}

func (p *OnlinePacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	onlineCount := len(p.UserService.GetLoggedCharacters())
	message := fmt.Sprintf("Hay %d usuarios conectados.", onlineCount)
	
	connection.Send(&outgoing.ConsoleMessagePacket{
		Message: message,
		Font:    outgoing.INFO,
	})
	
	return true, nil
}
