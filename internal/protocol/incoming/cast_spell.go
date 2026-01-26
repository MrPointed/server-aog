package incoming

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/service"
)

type CastSpellPacket struct {
	MapService   service.MapService
	SpellService service.SpellService
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

	// Update selected spell
	char.SelectedSpell = spellID
	
	connection.Send(&outgoing.ConsoleMessagePacket{
		Message: "Hechizo seleccionado.",
		Font:    outgoing.INFO,
	})

	return true, nil
}
