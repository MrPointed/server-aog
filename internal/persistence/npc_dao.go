package persistence

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ao-go-server/internal/model"
)

type NpcDAO struct {
	path string
}

func NewNpcDAO(path string) *NpcDAO {
	return &NpcDAO{path: path}
}

func (d *NpcDAO) Load() (map[int]*model.NPC, error) {
	data, err := ReadINI(d.path)
	if err != nil {
		return nil, err
	}

	npcs := make(map[int]*model.NPC)

	for section, props := range data {
		if !strings.HasPrefix(section, "NPC") || section == "NPC_COUNT" {
			continue
		}

		id, err := strconv.Atoi(section[3:])
		if err != nil {
			continue
		}

		npc := &model.NPC{
			ID:          id,
			Name:        props["NAME"],
			Description: props["DESCRIPTION"],
			Type:        model.NPCType(toInt(props["NPC_TYPE"])),
			Head:        toInt(props["HEAD"]),
			Body:        toInt(props["BODY"]),
			Heading:     model.Heading(toInt(props["HEADING"]) - 1),
			Level:       toInt(props["LEVEL"]),
			Exp:         toInt(props["EXP"]),
			MinHit:      toInt(props["MIN_HIT"]),
			MaxHit:      toInt(props["MAX_HIT"]),
			Hostile:     props["HOSTILE"] == "1",
			Movement:    props["MOVEMENT"] == "1",
		}

		// HP can be MinHP/MaxHP or just HP
		if hp, ok := props["MAX_HP"]; ok {
			npc.MaxHp = toInt(hp)
			npc.Hp = toInt(props["MIN_HP"])
		} else if hp, ok := props["HP"]; ok {
			npc.MaxHp = toInt(hp)
			npc.Hp = npc.MaxHp
		}

		if npc.Heading < 0 {
			npc.Heading = model.South
		}

		// Load drops
		for i := 1; i <= 10; i++ {
			dropKey := fmt.Sprintf("DROP%d", i)
			if val, ok := props[dropKey]; ok {
				parts := strings.Split(val, "-")
				if len(parts) == 2 {
					npc.Drops = append(npc.Drops, model.NPCDrop{
						ObjectID: toInt(parts[0]),
						Amount:   toInt(parts[1]),
					})
				}
			}
		}

		npcs[id] = npc
	}

	return npcs, nil
}
