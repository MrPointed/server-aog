package incoming

import (
	"fmt"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
)

type DoubleClickPacket struct{}

func (p *DoubleClickPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	if buffer.ReadableBytes() < 2 {
		return false, nil
	}

	rawX, _ := buffer.Get()
	rawY, _ := buffer.Get()

	x := rawX - 1
	y := rawY - 1

	char := connection.GetUser()
	name := "Unknown"
	if char != nil {
		name = char.Name
	}

	fmt.Printf("DOUBLE_CLICK from [%s] at X:%d Y:%d\n", name, x, y)

	return true, nil
}

