package incoming

import (
	"fmt"

	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/service"
)

type PickUpPacket struct {
	MapService     *service.MapService
	MessageService *service.MessageService
}

func (p *PickUpPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	char := connection.GetUser()
	if char == nil {
		return true, nil
	}

	worldObj := p.MapService.GetObjectAt(char.Position)
	if worldObj == nil {
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "No hay nada en el suelo.",
			Font:    outgoing.INFO,
		})
		return true, nil
	}

	if !worldObj.Object.Pickupable {
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "No puedes levantar este objeto.",
			Font:    outgoing.INFO,
		})
		return true, nil
	}

	// Add to inventory
	if char.Inventory.AddItem(worldObj.Object.ID, worldObj.Amount) {
		// Sync inventory (simplified: update all slots or just the modified one?)
		// Find which slot was used (AddItem should return it)
		// For now, let's just find the slot with this ID
		for i := 0; i < 30; i++ {
			slot := char.Inventory.GetSlot(i)
			if slot.ObjectID == worldObj.Object.ID {
				connection.Send(&outgoing.ChangeInventorySlotPacket{
					Slot:     byte(i + 1),
					Object:   worldObj.Object,
					Amount:   slot.Amount,
					Equipped: slot.Equipped,
				})
				break
			}
		}

		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: fmt.Sprintf("Has levantado %d %s.", worldObj.Amount, worldObj.Object.Name),
			Font:    outgoing.INFO,
		})

		// Remove from map
		p.MapService.RemoveObject(char.Position)
		p.MessageService.SendToArea(&outgoing.ObjectDeletePacket{
			X: char.Position.X,
			Y: char.Position.Y,
		}, char.Position)

	} else {
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: "No tienes espacio en el inventario.",
			Font:    outgoing.INFO,
		})
	}

	return true, nil
}
