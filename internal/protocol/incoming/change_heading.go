package incoming

import (
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/service"
)

type ChangeHeadingPacket struct {
	AreaService service.AreaService
}

func (p *ChangeHeadingPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	if buffer.ReadableBytes() < 1 {
		return false, nil
	}

	headingByte, _ := buffer.Get()
	heading := model.Heading(headingByte - 1)

	// Validate heading
	if heading < model.North || heading > model.West {
		return true, nil
	}

	user := connection.GetUser()
	if user == nil {
		return true, nil
	}

	//if user's paralyzed can't head
	if user.Paralyzed {
		return true, nil
	}

	if user.Heading != heading {
		user.Heading = heading
		// Notify nearby players about the heading change
		p.AreaService.BroadcastToArea(user.Position, &outgoing.CharacterChangePacket{Character: user})
	}

	return true, nil
}
