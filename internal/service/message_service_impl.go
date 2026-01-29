package service

import (
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type MessageServiceImpl struct {
	userService   UserService
	areaService   AreaService
	mapService    MapService
	objectService ObjectService
}

func NewMessageServiceImpl(userService UserService, areaService AreaService, mapService MapService, objectService ObjectService) MessageService {
	return &MessageServiceImpl{
		userService:   userService,
		areaService:   areaService,
		mapService:    mapService,
		objectService: objectService,
	}
}

func (s *MessageServiceImpl) MapService() MapService {
	return s.mapService
}

func (s *MessageServiceImpl) UserService() UserService {
	return s.userService
}

func (s *MessageServiceImpl) AreaService() AreaService {
	return s.areaService
}

func (s *MessageServiceImpl) HandleDeath(char *model.Character, message string) {
	if char.Dead {
		return
	}

	isSafe := s.mapService.IsSafeZone(char.Position)
	isPkMap := s.mapService.IsPkMap(char.Position.Map)
	shouldDropItems := !isSafe && isPkMap

	char.Dead = true
	char.Hp = 0
	char.Poisoned = false
	char.Paralyzed = false
	char.Immobilized = false
	char.Body = 8   // Ghost
	char.Head = 500 // Ghost head
	char.Weapon = 0
	char.Shield = 0
	char.Helmet = 0
	char.SetStateChanged()

	conn := s.userService.GetConnection(char)

	// Process inventory
	for i := 0; i < model.InventorySlots; i++ {
		slot := char.Inventory.GetSlot(i)
		if slot.ObjectID == 0 {
			continue
		}

		obj := s.objectService.GetObject(slot.ObjectID)
		if obj == nil {
			continue
		}

		// Drop logic
		if shouldDropItems && !obj.Newbie && !obj.NoDrop {
			dropPos := s.findDropPosition(char.Position)
			if dropPos != nil {
				worldObj := &model.WorldObject{
					Object: obj,
					Amount: slot.Amount,
				}
				s.mapService.PutObject(*dropPos, worldObj)

				// Notify nearby players about the new object on ground
				s.SendToArea(&outgoing.ObjectCreatePacket{
					X:            dropPos.X,
					Y:            dropPos.Y,
					GraphicIndex: int16(obj.GraphicIndex),
				}, *dropPos)
			}

			// Remove from inventory
			slot.ObjectID = 0
			slot.Amount = 0
			slot.Equipped = false
		} else {
			// Keep item but unequip it
			if slot.Equipped {
				slot.Equipped = false
			}
		}
	}

	if conn != nil {
		conn.Send(outgoing.NewUpdateUserStatsPacket(char))
		if message != "" {
			conn.Send(&outgoing.ConsoleMessagePacket{
				Message: message,
				Font:    outgoing.INFO,
			})
		} else {
			conn.Send(&outgoing.ConsoleMessagePacket{
				Message: "¡Has muerto!",
				Font:    outgoing.INFO,
			})
		}
		// Sync full inventory to reflect dropped items
		for i := 0; i < model.InventorySlots; i++ {
			slot := char.Inventory.GetSlot(i)
			var obj *model.Object
			if slot.ObjectID > 0 {
				obj = s.objectService.GetObject(slot.ObjectID)
			}

			conn.Send(&outgoing.ChangeInventorySlotPacket{
				Slot:     byte(i + 1),
				Object:   obj,
				Amount:   slot.Amount,
				Equipped: slot.Equipped,
			})
		}
	}

	// Broadcast character appearance change (ghost)
	s.SendToArea(&outgoing.CharacterChangePacket{Character: char}, char.Position)
}

func (s *MessageServiceImpl) checkDropPos(center model.Position, dx, dy int) *model.Position {
	tx := int(center.X) + dx
	ty := int(center.Y) + dy

	// Check strictly playable area from MapService IsBlocked/IsInPlayableArea
	if !s.mapService.IsInPlayableArea(tx, ty) {
		return nil
	}
	
	if s.mapService.IsBlocked(center.Map, tx, ty) {
		return nil
	}

	pos := model.Position{X: byte(tx), Y: byte(ty), Map: center.Map}
	if s.mapService.GetObjectAt(pos) == nil {
		return &pos
	}
	return nil
}

func (s *MessageServiceImpl) findDropPosition(startPos model.Position) *model.Position {
	// 1. Check center
	if pos := s.checkDropPos(startPos, 0, 0); pos != nil {
		return pos
	}

	// 2. Spiral out
	for r := 1; r <= 3; r++ {
		// Iterate around the square ring of radius r
		for i := -r; i <= r; i++ {
			// Top row (y = -r)
			if pos := s.checkDropPos(startPos, i, -r); pos != nil {
				return pos
			}
			// Bottom row (y = r)
			if pos := s.checkDropPos(startPos, i, r); pos != nil {
				return pos
			}
			// Left col (x = -r)
			if pos := s.checkDropPos(startPos, -r, i); pos != nil {
				return pos
			}
			// Right col (x = r)
			if pos := s.checkDropPos(startPos, r, i); pos != nil {
				return pos
			}
		}
	}
	return nil
}

func (s *MessageServiceImpl) HandleResurrection(char *model.Character) {
	if !char.Dead {
		return
	}

	char.Dead = false
	char.Hp = char.MaxHp
	char.Head = char.OriginalHead
	char.Body = s.userService.BodyService().GetBody(char.Race, char.Gender)
	char.SetStateChanged()

	conn := s.userService.GetConnection(char)
	if conn != nil {
		conn.Send(&outgoing.ConsoleMessagePacket{
			Message: "Has sido resucitado.",
			Font:    outgoing.INFO,
		})
		// Sync self
		conn.Send(outgoing.NewUpdateUserStatsPacket(char))
	}

	// Broadcast change
	s.SendToArea(&outgoing.CharacterChangePacket{Character: char}, char.Position)
}

func (s *MessageServiceImpl) SendMessage(char *model.Character, msg string, msgType outgoing.Font) {
	s.SendConsoleMessage(char, msg, msgType)
}

func (s *MessageServiceImpl) SendConsoleMessage(user *model.Character, message string, font outgoing.Font) {
	conn := s.userService.GetConnection(user)
	if conn != nil {
		conn.Send(&outgoing.ConsoleMessagePacket{
			Message: message,
			Font:    font,
		})
	}
}

func (s *MessageServiceImpl) BroadcastMessage(msg string, msgType outgoing.Font) {
	s.SendToAll(&outgoing.ConsoleMessagePacket{Message: msg, Font: msgType})
}

func (s *MessageServiceImpl) BroadcastNearby(char *model.Character, packet protocol.OutgoingPacket) {
	s.areaService.BroadcastNearby(char, packet)
}

func (s *MessageServiceImpl) SendToAll(packet protocol.OutgoingPacket) {
	for _, char := range s.userService.GetLoggedCharacters() {
		conn := s.userService.GetConnection(char)
		if conn != nil {
			conn.Send(packet)
		}
	}
}

func (s *MessageServiceImpl) SendToArea(packet protocol.OutgoingPacket, pos model.Position) {
	s.areaService.BroadcastToArea(pos, packet)
}

func (s *MessageServiceImpl) SendToAreaButUser(packet protocol.OutgoingPacket, pos model.Position, exclude *model.Character) {
	s.mapService.ForEachCharacter(pos.Map, func(char *model.Character) {
		if char != exclude {
			if s.areaService.InRange(pos, char.Position) {
				conn := s.userService.GetConnection(char)
				if conn != nil {
					conn.Send(packet)
				}
			}
		}
	})
}

func (s *MessageServiceImpl) SendToMap(packet protocol.OutgoingPacket, mapId int) {
	s.mapService.ForEachCharacter(mapId, func(char *model.Character) {
		conn := s.userService.GetConnection(char)
		if conn != nil {
			conn.Send(packet)
		}
	})
}