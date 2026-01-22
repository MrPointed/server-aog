package persistence

import (
	"strconv"
	"strings"

	"github.com/ao-go-server/internal/model"
)

type SpellDAO struct {
	path string
}

func NewSpellDAO(path string) *SpellDAO {
	return &SpellDAO{path: path}
}

func (d *SpellDAO) Load() (map[int]*model.Spell, error) {
	data, err := ReadINI(d.path)
	if err != nil {
		return nil, err
	}

	spells := make(map[int]*model.Spell)

	for section, props := range data {
		if !strings.HasPrefix(section, "HECHIZO") || section == "INIT" {
			continue
		}

		id, err := strconv.Atoi(section[7:])
		if err != nil {
			continue
		}

		spell := &model.Spell{
			ID:              id,
			Name:            props["NOMBRE"],
			Description:     props["DESC"],
			MagicWords:      props["PALABRASMAGICAS"],
			CasterMsg:       props["HECHIZEROMSG"],
			OwnMsg:          props["PROPIOMSG"],
			TargetMsg:       props["TARGETMSG"],
			Type:            toInt(props["TIPO"]),
			WAV:             toInt(props["WAV"]),
			FX:              toInt(props["FXGRH"]),
			Loops:           toInt(props["LOOPS"]),
			MinSkill:        toInt(props["MINSKILL"]),
			ManaRequired:    toInt(props["MANAREQUERIDO"]),
			StaminaRequired: toInt(props["STAREQUERIDO"]),
			TargetType:      model.SpellTarget(toInt(props["TARGET"])),
			MinHP:           toInt(props["MINHP"]),
			MaxHP:           toInt(props["MAXHP"]),
			Invisibility:    props["INVISIBILIDAD"] == "1",
			Paralyzes:       props["PARALIZA"] == "1",
			Immobilizes:     props["INMOVILIZA"] == "1",
		}

		spells[id] = spell
	}

	return spells, nil
}
