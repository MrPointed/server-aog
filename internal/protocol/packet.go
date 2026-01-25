package protocol

import (
	"time"

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
	GetRemoteAddr() string
	GetStats() (in uint64, out uint64, pIn uint64, pOut uint64, start time.Time)
}
