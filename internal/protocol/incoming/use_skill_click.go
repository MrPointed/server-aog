package incoming

import (
	"fmt"
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/service"
)

type UseSkillClickPacket struct {
	SkillService service.SkillService
}

func (p *UseSkillClickPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	// Packet format: X (byte), Y (byte), Skill (byte)
	if buffer.ReadableBytes() < 3 {
		return false, nil
	}

	x, _ := buffer.Get()
	y, _ := buffer.Get()
	skillByte, _ := buffer.Get()
	
	skill := model.Skill(skillByte)

	// Adjust coordinates (Client sends 1-based usually, need to check. VB6 code: just ReadByte. Usually AO coords are 1-based in packets but 0-based in server arrays. MapService expects 0-based? Wait.
	// In double_click.go: x := rawX - 1.
	// In walk.go: usually raw.
	// Standard AO: Packets send X, Y. If map array is 1-100, it's direct.
	// If Go map array is 0-99, we subtract 1.
	// Go server 'double_click.go' subtracts 1. So we should too.
	
	mapX := x - 1
	mapY := y - 1

	user := connection.GetUser()
	if user != nil {
		fmt.Printf("UseSkillClick: User %s, Skill %d at %d,%d\n", user.Name, skill, mapX, mapY)
		p.SkillService.HandleUseSkillClick(user, skill, mapX, mapY)
	}

	return true, nil
}