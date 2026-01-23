package outgoing

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/model"
)

type UpdateUserStatsPacket struct {
	MaxHp       int
	Hp          int
	MaxMana     int
	Mana        int
	MaxStamina  int
	Stamina     int
	Money       int
	Level       byte
	ExpToNext   int
	Exp         int
	SkillPoints int
}

func NewUpdateUserStatsPacket(char *model.Character) *UpdateUserStatsPacket {
	return &UpdateUserStatsPacket{
		MaxHp:       char.MaxHp,
		Hp:          char.Hp,
		MaxMana:     char.MaxMana,
		Mana:        char.Mana,
		MaxStamina:  char.MaxStamina,
		Stamina:     char.Stamina,
		Money:       char.Gold,
		Level:       char.Level,
		ExpToNext:   char.ExpToNext,
		Exp:         char.Exp,
		SkillPoints: char.SkillPoints,
	}
}

func (p *UpdateUserStatsPacket) Write(buffer *network.DataBuffer) error {
	buffer.PutShort(int16(p.MaxHp))
	buffer.PutShort(int16(p.Hp))
	buffer.PutShort(int16(p.MaxMana))
	buffer.PutShort(int16(p.Mana))
	buffer.PutShort(int16(p.MaxStamina))
	buffer.PutShort(int16(p.Stamina))
	buffer.PutInt(int32(p.Money))
	buffer.Put(p.Level)
	buffer.PutInt(int32(p.ExpToNext))
	buffer.PutInt(int32(p.Exp))
	buffer.PutInt(int32(p.SkillPoints))
	return nil
}
