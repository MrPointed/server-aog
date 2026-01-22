package service

import (
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type MessageService struct {
	userService *UserService
	AreaService *AreaService
	MapService  *MapService
}

func NewMessageService(userService *UserService, areaService *AreaService, mapService *MapService) *MessageService {
	return &MessageService{
		userService: userService,
		AreaService: areaService,
		MapService:  mapService,
	}
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
	gameMap := s.MapService.GetMap(mapId)
	if gameMap == nil {
		return
	}
	for _, char := range gameMap.Characters {
		conn := s.userService.GetConnection(char)
		if conn != nil {
			conn.Send(packet)
		}
	}
}

func (s *MessageService) SendToMapButUser(packet protocol.OutgoingPacket, mapId int, exclude *model.Character) {
	gameMap := s.MapService.GetMap(mapId)
	if gameMap == nil {
		return
	}
	for _, char := range gameMap.Characters {
		if char != exclude {
			conn := s.userService.GetConnection(char)
			if conn != nil {
				conn.Send(packet)
			}
		}
	}
}

func (s *MessageService) SendToArea(packet protocol.OutgoingPacket, pos model.Position) {
	s.AreaService.BroadcastToArea(pos, packet)
}

func (s *MessageService) SendToAreaButUser(packet protocol.OutgoingPacket, pos model.Position, exclude *model.Character) {
	gameMap := s.MapService.GetMap(pos.Map)
	if gameMap == nil {
		return
	}

	for _, char := range gameMap.Characters {
		if char != exclude {
			if s.AreaService.InRange(pos, char.Position) {
				conn := s.userService.GetConnection(char)
				if conn != nil {
					conn.Send(packet)
				}
			}
		}
	}
}
