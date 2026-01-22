package incoming

import (
	"fmt"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
)

type WorkLeftClickPacket struct{}

func (p *WorkLeftClickPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	if buffer.ReadableBytes() < 3 {
		return false, nil
	}

	rawX, _ := buffer.Get()
	rawY, _ := buffer.Get()
	skill, _ := buffer.Get()

	x := rawX - 1
	y := rawY - 1

	char := connection.GetUser()
	name := "Unknown"
	if char != nil {
		name = char.Name
	}

	fmt.Printf("WORK_LEFT_CLICK from [%s] at X:%d Y:%d with Skill:%d\n", name, x, y, skill)

	return true, nil
}

