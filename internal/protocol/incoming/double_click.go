package incoming

import (
	"fmt"
	"math"

	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/service"
)

type DoubleClickPacket struct {
	MapService    service.MapService
	NpcService    service.NpcService
	UserService   service.UserService
	ObjectService service.ObjectService
	AreaService   service.AreaService
	BankService   service.BankService
	SpellService  service.SpellService
}

func (p *DoubleClickPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	if buffer.ReadableBytes() < 2 {
		return false, nil
	}

	rawX, _ := buffer.Get()
	rawY, _ := buffer.Get()

	x := rawX - 1
	y := rawY - 1

	user := connection.GetUser()
	if user != nil {
		fmt.Printf("DoubleClick from %s at %d,%d (Raw: %d,%d)\n", user.Name, x, y, rawX, rawY)
	}
	if user == nil {
		return true, nil
	}

	// Vision Range Check
	const RANGO_VISION_X = 8
	const RANGO_VISION_Y = 6

	if math.Abs(float64(user.Position.X)-float64(x)) > RANGO_VISION_X ||
		math.Abs(float64(user.Position.Y)-float64(y)) > RANGO_VISION_Y {
		return true, nil
	}

	mapID := user.Position.Map
	gameMap := p.MapService.GetMap(mapID)
	if gameMap == nil {
		return true, nil
	}

	// Helper distance function
	getDist := func(p1 model.Position, tx, ty int) int {
		return int(math.Abs(float64(p1.X)-float64(tx)) + math.Abs(float64(p1.Y)-float64(ty)))
	}

	tile := gameMap.GetTile(int(x), int(y))

	// 1. Check NPCs
	if tile.NPC != nil {
		npc := tile.NPC
		user.TargetNPC = npc.Index
		user.TargetUser = 0
		user.TargetNpcType = npc.NPC.Type

		dist := getDist(user.Position, int(x), int(y))

		// Show NPC description if available
		if npc.NPC.Description != "" {
			fmt.Printf("NPC DoubleClick: Sending description for NPC %d: %s\n", npc.NPC.ID, npc.NPC.Description)
			connection.Send(&outgoing.ConsoleMessagePacket{
				Message: fmt.Sprintf("%s: %s", npc.NPC.Name, npc.NPC.Description),
				Font:    outgoing.INFO,
			})

			// Show as overhead text
			p.AreaService.BroadcastToArea(npc.Position, &outgoing.ChatOverHeadPacket{
				Message:   npc.NPC.Description,
				CharIndex: npc.Index,
				R:         255,
				G:         255,
				B:         255,
			})
		} else {
			fmt.Printf("NPC DoubleClick: NPC %d has no description\n", npc.NPC.ID)
		}

		switch npc.NPC.Type {
		case model.NTCommon:
			if npc.NPC.CanTrade {
				if user.Dead {
					connection.Send(&outgoing.ConsoleMessagePacket{Message: "¡Estás muerto!", Font: outgoing.INFO})
					return true, nil
				}
				if dist > 3 {
					connection.Send(&outgoing.ConsoleMessagePacket{Message: "Estás demasiado lejos del vendedor.", Font: outgoing.INFO})
					return true, nil
				}

				user.TradingNPCIndex = npc.Index
				connection.Send(&outgoing.CommerceInitPacket{})
				
				// Send NPC Inventory
				for i, slot := range npc.NPC.Inventory {
					obj := p.ObjectService.GetObject(slot.ObjectID)
					if obj != nil {
						connection.Send(&outgoing.ChangeNpcInventorySlotPacket{
							Slot:   byte(i + 1),
							Object: obj,
							Amount: slot.Amount,
						})
					}
				}
				return true, nil
			}
			connection.Send(&outgoing.ConsoleMessagePacket{Message: "Este personaje no tiene nada para venderte.", Font: outgoing.INFO})

		case model.NTBanker:
			if user.Dead {
				connection.Send(&outgoing.ConsoleMessagePacket{Message: "¡Estás muerto!", Font: outgoing.INFO})
				return true, nil
			}
			if dist > 3 {
				connection.Send(&outgoing.ConsoleMessagePacket{Message: "Estás demasiado lejos del banquero.", Font: outgoing.INFO})
				return true, nil
			}
			p.BankService.OpenBank(user)
			return true, nil

		case model.NTHealer, model.NTHealerNewbie:
			if dist > 10 {
				connection.Send(&outgoing.ConsoleMessagePacket{Message: "El sacerdote no puede curarte debido a que estás demasiado lejos.", Font: outgoing.INFO})
				return true, nil
			}

			if user.Dead {
				p.SpellService.SacerdoteResucitateUser(user)
			} else if user.Hp < user.MaxHp {
				p.SpellService.SacerdoteHealUser(user)
			}
		}
		return true, nil
	}

	// 2. Check Objects (Doors, Signs, etc.)
	var targetObj *model.WorldObject
	tx, ty := int(x), int(y)

	findObject := func() {
		if obj := gameMap.GetTile(tx, ty).Object; obj != nil {
			targetObj = obj
			return
		}
		// Check adjacent for Doors (as per VB6 logic)
		offsets := []struct{ dx, dy int }{{1, 0}, {1, 1}, {0, 1}}
		for _, off := range offsets {
			nx, ny := int(x)+off.dx, int(y)+off.dy
			if nx >= 0 && nx < model.MapWidth && ny >= 0 && ny < model.MapHeight {
				obj := gameMap.GetTile(nx, ny).Object
				if obj != nil && obj.Object.Type == model.OTDoor {
					targetObj = obj
					tx, ty = nx, ny
					return
				}
			}
		}
	}

	findObject()

	if targetObj != nil {
		user.TargetObj = targetObj.Object.ID

		dist := getDist(user.Position, tx, ty)
		if dist > 2 {
			connection.Send(&outgoing.ConsoleMessagePacket{Message: "Estás demasiado lejos.", Font: outgoing.INFO})
			return true, nil
		}

		switch targetObj.Object.Type {
		case model.OTDoor:
			tile := gameMap.GetTile(tx, ty)
			// Toggle Door
			newObjID := 0
			shouldBlock := false

			// Determine action based on Blocked state + Index availability
			// Prefer Closing if not blocked and has ClosedIndex
			// Prefer Opening if blocked and has OpenIndex

			if !tile.Blocked && targetObj.Object.ClosedIndex != 0 {
				// Close
				newObjID = targetObj.Object.ClosedIndex
				shouldBlock = true
			} else if tile.Blocked && targetObj.Object.OpenIndex != 0 {
				// Open
				newObjID = targetObj.Object.OpenIndex
				shouldBlock = false
			} else {
				// Fallback if blocked status is desynced with object type
				if targetObj.Object.OpenIndex != 0 {
					newObjID = targetObj.Object.OpenIndex
					shouldBlock = false
				} else if targetObj.Object.ClosedIndex != 0 {
					newObjID = targetObj.Object.ClosedIndex
					shouldBlock = true
				}
			}

			if newObjID != 0 {
				newDef := p.ObjectService.GetObject(newObjID)
				if newDef != nil {
					targetObj.Object = newDef
					tile.Blocked = shouldBlock

					// Update left tile as well (Doors are 2-wide usually)
					if tx > 0 {
						leftTile := gameMap.GetTile(tx-1, ty)
						leftTile.Blocked = shouldBlock
						p.AreaService.BroadcastToArea(model.Position{X: byte(tx - 1), Y: byte(ty), Map: mapID}, &outgoing.BlockPositionPacket{
							X:       byte(tx - 1),
							Y:       byte(ty),
							Blocked: shouldBlock,
						})
					}

					// Broadcast visual change to clients in area
					p.AreaService.BroadcastToArea(model.Position{X: byte(tx), Y: byte(ty), Map: mapID}, &outgoing.ObjectCreatePacket{
						X:            byte(tx),
						Y:            byte(ty),
						GraphicIndex: int16(newDef.GraphicIndex),
					})

					// Update blocking status on clients
					p.AreaService.BroadcastToArea(model.Position{X: byte(tx), Y: byte(ty), Map: mapID}, &outgoing.BlockPositionPacket{
						X:       byte(tx),
						Y:       byte(ty),
						Blocked: shouldBlock,
					})

					// Play Sound
					p.AreaService.BroadcastToArea(model.Position{X: byte(tx), Y: byte(ty), Map: mapID}, &outgoing.PlayWavePacket{
						Wave: 9,
						X:    byte(tx),
						Y:    byte(ty),
					})
				}
			}

		case model.OTSign:
			// TODO: Send WriteShowSignal(user, targetObj.Object.ID)
			connection.Send(&outgoing.ConsoleMessagePacket{Message: "Lees el cartel...", Font: outgoing.INFO})
		}
	}

	return true, nil
}
