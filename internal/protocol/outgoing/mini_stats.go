package outgoing

import (
	"github.com/ao-go-server/internal/network"
)

type MiniStatsPacket struct {
	CriminalsKilled int32
	CitizensKilled  int32
	UsersKilled     int32
	CreaturesKilled int16
	Role            byte
	JailTime        int32
}

func (p *MiniStatsPacket) Write(buffer *network.DataBuffer) error {
	buffer.PutInt(p.CriminalsKilled)
	buffer.PutInt(p.CitizensKilled)
	buffer.PutInt(p.UsersKilled)
	buffer.PutShort(p.CreaturesKilled)
	buffer.Put(p.Role)
	buffer.PutInt(p.JailTime)
	return nil
}
