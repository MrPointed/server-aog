package incoming

import (
	"fmt"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/service"
)



type LoginNewCharacterPacket struct {

	LoginService *service.LoginService

}



func (p *LoginNewCharacterPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	initialPos := buffer.Pos()

	// Read Nick
	nick, err := buffer.GetUTF8String()
	if err != nil {
		return false, nil
	}

	// Read Password
	password, err := buffer.GetUTF8String()
	if err != nil {
		buffer.SetPos(initialPos)
		return false, nil
	}

	// Check for version and other fields
	// Version (3) + Race (1) + Gender (1) + Archetype (1) + Head (2) = 8 bytes
	if buffer.ReadableBytes() < 3+1+1+1+2 {
		buffer.SetPos(initialPos)
		return false, nil
	}

	v1, _ := buffer.Get()
	v2, _ := buffer.Get()
	v3, _ := buffer.Get()
	version := fmt.Sprintf("%d.%d.%d", v1, v2, v3)
	clientHash := ""

	raceId, _ := buffer.Get()
	genderId, _ := buffer.Get()
	archetypeId, _ := buffer.Get()
	headIdShort, _ := buffer.GetShort()
	headId := byte(headIdShort)

	// Read Mail
	mail, err := buffer.GetUTF8String()
	if err != nil {
		buffer.SetPos(initialPos)
		return false, nil
	}

	// CityId
	if buffer.ReadableBytes() < 1 {
		buffer.SetPos(initialPos)
		return false, nil
	}
	cityId, _ := buffer.Get()

	err = p.LoginService.ConnectNewCharacter(connection, nick, password, mail, raceId, genderId, archetypeId, headId, cityId, clientHash, version)

	if err != nil {

		connection.Send(&outgoing.ErrorMessagePacket{Message: err.Error()})

		connection.Disconnect()

		return true, nil

	}



	return true, nil

}
