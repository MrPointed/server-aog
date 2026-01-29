package service

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type GmServiceImpl struct {
	userService    UserService
	mapService     MapService
	messageService MessageService
	loginService   LoginService
}

func NewGmServiceImpl(userService UserService, mapService MapService, messageService MessageService, loginService LoginService) *GmServiceImpl {
	return &GmServiceImpl{
		userService:    userService,
		mapService:     mapService,
		messageService: messageService,
		loginService:   loginService,
	}
}

func (s *GmServiceImpl) HandleCommand(conn protocol.Connection, cmdID byte, buffer *network.DataBuffer) (bool, error) {
	user := conn.GetUser()
	if user == nil || !user.Privileges.IsGM() {
		return true, nil
	}

	switch cmdID {
	case 1: // /GMSG
		return s.handleGMMessage(user, buffer)
	case 2: // /SHOWNAME
		return s.handleShowName(conn, user)
	case 7: // /HORA (ServerTime)
		return s.handleServerTime(conn)
	case 11: // /TELEP (WarpChar)
		return s.handleWarpChar(conn, buffer)
	case 15: // /IRA (GoToChar)
		return s.handleGoToChar(conn, buffer)
	case 32: // /ONLINEGM
		return s.handleOnlineGM(conn)
	case 33: // /DOBACKUP
		return s.handleDoBackup(conn)
	default:
		slog.Warn("GM sent unknown command", "gm", user.Name, "command_id", cmdID)
	}
	return true, nil
}

func (s *GmServiceImpl) handleDoBackup(conn protocol.Connection) (bool, error) {
	s.loginService.WorldSave()
	return true, nil
}

func (s *GmServiceImpl) handleGMMessage(sender *model.Character, buffer *network.DataBuffer) (bool, error) {
	msg, err := buffer.GetUTF8String()
	if err != nil {
		return false, nil
	}

	packet := &outgoing.ConsoleMessagePacket{
		Message: fmt.Sprintf("%s> %s", sender.Name, msg),
		Font:    outgoing.GMMSG,
	}

	for _, char := range s.userService.GetLoggedCharacters() {
		if char.Privileges.IsGM() {
			conn := s.userService.GetConnection(char)
			if conn != nil {
				conn.Send(packet)
			}
		}
	}

	return true, nil
}

func (s *GmServiceImpl) handleShowName(conn protocol.Connection, user *model.Character) (bool, error) {
	msg := "Name visibility toggled."
	conn.Send(&outgoing.ConsoleMessagePacket{Message: msg, Font: outgoing.INFO})
	return true, nil
}

func (s *GmServiceImpl) handleServerTime(conn protocol.Connection) (bool, error) {
	msg := fmt.Sprintf("Hora del servidor: %s", time.Now().Format("15:04:05"))
	conn.Send(&outgoing.ConsoleMessagePacket{Message: msg, Font: outgoing.INFO})
	return true, nil
}

func (s *GmServiceImpl) handleWarpChar(conn protocol.Connection, buffer *network.DataBuffer) (bool, error) {
	targetName, err := buffer.GetUTF8String()
	if err != nil {
		return false, nil
	}

	mapID, _ := buffer.GetShort() // int16
	x, _ := buffer.Get()
	y, _ := buffer.Get()

	var targetChar *model.Character
	if strings.EqualFold(targetName, "YO") {
		targetChar = conn.GetUser()
	} else {
		targetChar = s.userService.GetCharacterByName(targetName)
	}

	if targetChar == nil {
		conn.Send(&outgoing.ConsoleMessagePacket{Message: "Usuario offline.", Font: outgoing.INFO})
		return true, nil
	}

	if !s.mapService.IsInPlayableArea(int(x), int(y)) {
		conn.Send(&outgoing.ConsoleMessagePacket{Message: "Posicion invalida", Font: outgoing.INFO})
		return true, nil
	}

	newPos := model.Position{Map: int(mapID), X: x, Y: y}

	// Notify old area (User leaving)
	s.messageService.SendToAreaButUser(&outgoing.CharacterRemovePacket{CharIndex: targetChar.CharIndex}, targetChar.Position, targetChar)

	s.mapService.PutCharacterAtPos(targetChar, newPos)

	targetConn := s.userService.GetConnection(targetChar)
	if targetConn != nil {
		targetConn.Send(&outgoing.ChangeMapPacket{MapId: int(mapID), Version: s.mapService.GetMap(int(mapID)).Version})
		targetConn.Send(&outgoing.CharacterCreatePacket{Character: targetChar})
		targetConn.Send(&outgoing.UserCharIndexInServerPacket{UserIndex: targetChar.CharIndex})
		targetConn.Send(&outgoing.AreaChangedPacket{Position: newPos})
		targetConn.Send(&outgoing.PosUpdatePacket{X: x, Y: y})

		// Sync new area state to user (NPCs, Objects, Users)
		s.messageService.AreaService().SendAreaState(targetChar)
	}

	// Notify new area (User entering)
	s.messageService.SendToAreaButUser(&outgoing.CharacterCreatePacket{Character: targetChar}, newPos, targetChar)

	// FX and Sound
	s.messageService.SendToArea(&outgoing.CreateFxPacket{CharIndex: targetChar.CharIndex, FxID: 1, Loops: 0}, newPos)
	s.messageService.SendToArea(&outgoing.PlayWavePacket{Wave: 2, X: x, Y: y}, newPos)

	conn.Send(&outgoing.ConsoleMessagePacket{Message: "Usuario transportado.", Font: outgoing.INFO})

	return true, nil
}

func (s *GmServiceImpl) handleGoToChar(conn protocol.Connection, buffer *network.DataBuffer) (bool, error) {
	targetName, err := buffer.GetUTF8String()
	if err != nil {
		return false, nil
	}

	targetChar := s.userService.GetCharacterByName(targetName)
	if targetChar == nil {
		conn.Send(&outgoing.ConsoleMessagePacket{Message: "Usuario offline.", Font: outgoing.INFO})
		return true, nil
	}

	user := conn.GetUser()
	newPos := targetChar.Position
	// Basic legal pos check (simplified: just next to user)
	if newPos.X < 90 {
		newPos.X += 1
	} else {
		newPos.X -= 1
	}

	if !s.mapService.IsInPlayableArea(int(newPos.X), int(newPos.Y)) {
		conn.Send(&outgoing.ConsoleMessagePacket{Message: "El usuario destino esta fuera de los limites permitidos.", Font: outgoing.INFO})
		return true, nil
	}

	// Notify old area
	s.messageService.SendToAreaButUser(&outgoing.CharacterRemovePacket{CharIndex: user.CharIndex}, user.Position, user)

	s.mapService.PutCharacterAtPos(user, newPos)

	conn.Send(&outgoing.ChangeMapPacket{MapId: newPos.Map, Version: s.mapService.GetMap(newPos.Map).Version})
	conn.Send(&outgoing.CharacterCreatePacket{Character: user})
	conn.Send(&outgoing.UserCharIndexInServerPacket{UserIndex: user.CharIndex})
	conn.Send(&outgoing.AreaChangedPacket{Position: newPos})
	conn.Send(&outgoing.PosUpdatePacket{X: newPos.X, Y: newPos.Y})

	// Sync new area state
	s.messageService.AreaService().SendAreaState(user)

	// Notify new area
	s.messageService.SendToAreaButUser(&outgoing.CharacterCreatePacket{Character: user}, newPos, user)

	// FX and Sound
	s.messageService.SendToArea(&outgoing.CreateFxPacket{CharIndex: user.CharIndex, FxID: 1, Loops: 0}, newPos)
	s.messageService.SendToArea(&outgoing.PlayWavePacket{Wave: 1, X: newPos.X, Y: newPos.Y}, newPos)

	conn.Send(&outgoing.ConsoleMessagePacket{Message: "Has sido transportado.", Font: outgoing.INFO})

	return true, nil
}

func (s *GmServiceImpl) handleOnlineGM(conn protocol.Connection) (bool, error) {
	count := 0
	conn.Send(&outgoing.ConsoleMessagePacket{Message: "GMs Online:", Font: outgoing.INFO})
	for _, char := range s.userService.GetLoggedCharacters() {
		if char.Privileges.IsGM() {
			count++
			conn.Send(&outgoing.ConsoleMessagePacket{Message: char.Name, Font: outgoing.INFO})
		}
	}
	conn.Send(&outgoing.ConsoleMessagePacket{Message: fmt.Sprintf("Total: %d", count), Font: outgoing.INFO})
	return true, nil
}
