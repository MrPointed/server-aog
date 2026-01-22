package incoming

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/service"
	"github.com/ao-go-server/internal/model"
)

type CastSpellPacket struct {
	MapService   *service.MapService
	SpellService *service.SpellService
}

func (p *CastSpellPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	if buffer.ReadableBytes() < 1 {
		return false, nil
	}

	slotIdx, _ := buffer.Get()
	slot := int(slotIdx) - 1

	char := connection.GetUser()
	if char == nil || char.Dead {
		return true, nil
	}

	if slot < 0 || slot >= len(char.Spells) {
		return true, nil
	}

	spellID := char.Spells[slot]

	// Find target in front
	targetPos := char.Position
	switch char.Heading {
	case model.North:
		targetPos.Y--
	case model.South:
		targetPos.Y++
	case model.East:
		targetPos.X++
	case model.West:
		targetPos.X--
	}

	gameMap := p.MapService.GetMap(targetPos.Map)
	if gameMap == nil {
		return true, nil
	}

	tile := gameMap.GetTile(int(targetPos.X), int(targetPos.Y))
	var target any
	if tile.Character != nil {
		target = tile.Character
	} else if tile.NPC != nil {
		target = tile.NPC
	} else {
		target = char // Self cast if no target in front
	}

	p.SpellService.CastSpell(char, spellID, target)

	return true, nil
}
