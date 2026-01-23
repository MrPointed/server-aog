package service

import (
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/protocol"
)

type ItemBehavior interface {
	Use(char *model.Character, slot int, obj *model.Object, connection protocol.Connection)
}

type EquipBehavior interface {
	ToggleEquip(char *model.Character, slot int, obj *model.Object, connection protocol.Connection)
}
