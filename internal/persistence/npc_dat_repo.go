package persistence

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ao-go-server/internal/model"
)

type NpcDatRepo struct {
	path string
}

func NewNpcDatRepo(path string) *NpcDatRepo {
	return &NpcDatRepo{path: path}
}

func (d *NpcDatRepo) Load() (map[int]*model.NPC, error) {
	data, err := ReadINI(d.path)
	if err != nil {
		return nil, err
	}

	npcs := make(map[int]*model.NPC)

	for section, props := range data {
		if !strings.HasPrefix(section, "NPC") || section == "NPC_COUNT" {
			continue
		}

		id, err := strconv.Atoi(strings.TrimSpace(section[3:]))
		if err != nil {
			continue
		}
		desc := props["DESCRIPTION"]
		if desc == "" {
			desc = props["DESC"]
		}

		npcTypeStr := props["NPC_TYPE"]
		if npcTypeStr == "" {
			npcTypeStr = props["NPCTYPE"]
		}

		exp := toInt(props["EXP"])
		if exp == 0 {
			exp = toInt(props["GIVEEXP"])
			if exp == 0 {
				exp = toInt(props["GIVE_EXP"])
			}
		}

		minHit := toInt(props["MINHIT"])
		if minHit == 0 {
			minHit = toInt(props["MIN_HIT"])
		}
		maxHit := toInt(props["MAXHIT"])
		if maxHit == 0 {
			maxHit = toInt(props["MAX_HIT"])
		}

		attackPower := toInt(props["PODERATAQUE"])
		if attackPower == 0 {
			attackPower = toInt(props["PODER_ATAQUE"])
		}

		evasionPower := toInt(props["PODEREVASION"])
		if evasionPower == 0 {
			evasionPower = toInt(props["PODER_EVASION"])
		}

		npc := &model.NPC{
			ID:          id,
			Name:        props["NAME"],
			Description: desc,
			Type:        model.NPCType(toInt(npcTypeStr)),
			Head:        toInt(props["HEAD"]),
			Body:        toInt(props["BODY"]),
			Heading:     model.Heading(toInt(props["HEADING"]) - 1),
			Level:       toInt(props["LEVEL"]),
			Exp:         exp,
			MinHit:      minHit,
			MaxHit:      maxHit,
			AttackPower:  attackPower,
			EvasionPower: evasionPower,
			Defense:      toInt(props["DEF"]),
			MagicDefense: toInt(props["DEFENSAMAGICA"]),
			Hostile:     props["HOSTILE"] == "1",
			CanTrade:    props["COMERCIA"] == "1",
			Movement:    toInt(props["MOVEMENT"]),
			Respawn:     props["RESPAWN"] == "1" || props["RE_SPAWN"] == "1",
			CastsSpells: toInt(props["LANZASPELLS"]),
			DoubleAttack:  props["ATACADOBLE"] == "1",
		}

		if npc.CastsSpells > 0 {
			for i := 1; i <= npc.CastsSpells; i++ {
				spellKey := fmt.Sprintf("SP%d", i)
				if val, ok := props[spellKey]; ok {
					npc.Spells = append(npc.Spells, toInt(val))
				}
			}
		}

		if npc.CanTrade {
			nroItems := toInt(props["NROITEMS"])
			for i := 1; i <= nroItems; i++ {
				itemKey := fmt.Sprintf("OBJ%d", i)
				if val, ok := props[itemKey]; ok {
					parts := strings.Split(val, "-")
					if len(parts) == 2 {
						npc.Inventory = append(npc.Inventory, model.InventorySlot{
							ObjectID: toInt(parts[0]),
							Amount:   toInt(parts[1]),
						})
					}
				}
			}
		}

		// HP can be MinHP/MaxHP or just HP
		if hp, ok := props["MAXHP"]; ok {
			npc.MaxHp = toInt(hp)
			npc.Hp = toInt(props["MINHP"])
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
