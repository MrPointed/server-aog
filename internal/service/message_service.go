package service

import (
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type MessageService struct {
	userService   *UserService
	AreaService   *AreaService
	MapService    *MapService
	ObjectService *ObjectService
}

func NewMessageService(userService *UserService, areaService *AreaService, mapService *MapService, objectService *ObjectService) *MessageService {
	return &MessageService{
		userService:   userService,
		AreaService:   areaService,
		MapService:    mapService,
		ObjectService: objectService,
	}
}

func (s *MessageService) HandleDeath(char *model.Character, message string) {
	if char.Dead {
		return
	}

	isSafe := s.MapService.IsSafeZone(char.Position)
	isPkMap := s.MapService.IsPkMap(char.Position.Map)
	shouldDropItems := !isSafe && isPkMap

	char.Dead = true
	char.Hp = 0
	char.Poisoned = false
	char.Paralyzed = false
	char.Immobilized = false
	char.Body = 8   // Ghost
	char.Head = 500 // Ghost head

	conn := s.userService.GetConnection(char)

	// Process inventory
	for i := 0; i < model.InventorySlots; i++ {
		slot := char.Inventory.GetSlot(i)
		if slot.ObjectID == 0 {
			continue
		}

		obj := s.ObjectService.GetObject(slot.ObjectID)
		if obj == nil {
			continue
		}

		// Drop logic
		if shouldDropItems && !obj.Newbie && !obj.NoDrop {
			// Try to drop on current position if empty, or nearby
			if s.MapService.GetObjectAt(char.Position) == nil {
				worldObj := &model.WorldObject{
					Object: obj,
					Amount: slot.Amount,
				}
				s.MapService.PutObject(char.Position, worldObj)
				
				// Notify nearby players about the new object on ground
				s.SendToArea(&outgoing.ObjectCreatePacket{
					X:            char.Position.X,
					Y:            char.Position.Y,
					GraphicIndex: int16(obj.GraphicIndex),
				}, char.Position)
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
				obj = s.ObjectService.GetObject(slot.ObjectID)
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

func (s *MessageService) HandleResurrection(char *model.Character) {
	if !char.Dead {
		return
	}

	char.Dead = false
	char.Hp = char.MaxHp
	char.Head = char.OriginalHead
	char.Body = s.userService.BodyService.GetBody(char.Race, char.Gender)

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

func (s *MessageService) SendConsoleMessage(user *model.Character, message string, font outgoing.Font) {
	conn := s.userService.GetConnection(user)
	if conn != nil {
		conn.Send(&outgoing.ConsoleMessagePacket{
			Message: message,
			Font:    font,
		})
	}
}

func (s *MessageService) SendToAll(packet protocol.OutgoingPacket) {
	for _, char := range s.userService.GetLoggedCharacters() {
		conn := s.userService.GetConnection(char)
		if conn != nil {
			conn.Send(packet)
		}
	}
}

func (s *MessageService) SendToAllButUser(packet protocol.OutgoingPacket, exclude *model.Character) {
	for _, char := range s.userService.GetLoggedCharacters() {
		if char != exclude {
			conn := s.userService.GetConnection(char)
			if conn != nil {
				conn.Send(packet)
			}
		}
	}
}

func (s *MessageService) SendToMap(packet protocol.OutgoingPacket, mapId int) {
	s.MapService.ForEachCharacter(mapId, func(char *model.Character) {
		conn := s.userService.GetConnection(char)
		if conn != nil {
			conn.Send(packet)
		}
	})
}

func (s *MessageService) SendToMapButUser(packet protocol.OutgoingPacket, mapId int, exclude *model.Character) {
	s.MapService.ForEachCharacter(mapId, func(char *model.Character) {
		if char != exclude {
			conn := s.userService.GetConnection(char)
			if conn != nil {
				conn.Send(packet)
			}
		}
	})
}

func (s *MessageService) SendToArea(packet protocol.OutgoingPacket, pos model.Position) {
	s.AreaService.BroadcastToArea(pos, packet)
}

func (s *MessageService) SendToAreaButUser(packet protocol.OutgoingPacket, pos model.Position, exclude *model.Character) {
	s.MapService.ForEachCharacter(pos.Map, func(char *model.Character) {
		if char != exclude {
			if s.AreaService.InRange(pos, char.Position) {
				conn := s.userService.GetConnection(char)
				if conn != nil {
					conn.Send(packet)
				}
			}
		}
	})
}
