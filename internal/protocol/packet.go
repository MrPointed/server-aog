package protocol

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/model"
)

type IncomingPacket interface {
	Handle(buffer *network.DataBuffer, connection Connection) (bool, error)
}

type OutgoingPacket interface {
	Write(buffer *network.DataBuffer) error
}

type Connection interface {
	Send(packet OutgoingPacket) error
	Disconnect()
	SetAttribute(attr int, value byte)
	GetAttribute(attr int) byte
	GetUser() *model.Character
	SetUser(user *model.Character)
}
