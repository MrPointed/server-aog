package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type Font int

const (
	TALK Font = iota
	FIGHT
	WARNING
	INFO
	INFOBOLD
	EXECUTION
	PARTY
	POISON
	GUILD
	SERVER
	GUILDMSG
	COUNCIL
	CHAOSCOUNCIL
	COUNCILSee
	CHAOSCOUNCILSee
	SENTINEL
	GMMSG
	GM
	CITIZEN
	CONSE
	GOD
)

type ConsoleMessagePacket struct {
	Message string
	Font    Font
}

func (p *ConsoleMessagePacket) Write(buffer *network.DataBuffer) error {
	buffer.PutUTF8String(p.Message)
	buffer.Put(byte(p.Font))
	return nil
}
