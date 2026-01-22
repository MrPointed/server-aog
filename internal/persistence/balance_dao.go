package persistence

import (
	"fmt"
	"strconv"

	"github.com/ao-go-server/internal/model"
)

type BalanceDAO struct {
	path string
}

func NewBalanceDAO(path string) *BalanceDAO {
	return &BalanceDAO{path: path}
}

func (d *BalanceDAO) Load() (map[model.UserArchetype]*model.ArchetypeModifiers, map[model.Race]*model.RaceModifiers, error) {
	data, err := ReadINI(d.path)
	if err != nil {
		return nil, nil, err
	}

	archetypeModifiers := make(map[model.UserArchetype]*model.ArchetypeModifiers)

raceModifiers := make(map[model.Race]*model.RaceModifiers)

	// Races
	if modRaza, ok := data["MODRAZA"]; ok {
	
races := []struct {
		name model.Race
		prefix string
	}{
			{model.Human, "HUMANO"},
			{model.Elf, "ELFO"},
			{model.DarkElf, "DROW"},
			{model.Gnome, "GNOMO"},
			{model.Dwarf, "ENANO"},
		}

		for _, r := range races {
			raceModifiers[r.name] = &model.RaceModifiers{
				Strength:     toInt(modRaza[r.prefix+"FUERZA"]),
				Dexterity:    toInt(modRaza[r.prefix+"AGILIDAD"]),
				Intelligence: toInt(modRaza[r.prefix+"INTELIGENCIA"]),
				Charisma:     toInt(modRaza[r.prefix+"CARISMA"]),
				Constitution: toInt(modRaza[r.prefix+"CONSTITUCION"]),
			}
		}
	}

	// Archetypes
	archetypes := []struct {
		id   model.UserArchetype
		name string
	}{
		{model.Mage, "MAGO"},
		{model.Cleric, "CLERIGO"},
		{model.Warrior, "GUERRERO"},
		{model.Assasin, "ASESINO"},
		{model.Thief, "LADRON"},
		{model.Bard, "BARDO"},
		{model.Druid, "DRUIDA"},
		{model.Bandit, "BANDIDO"},
		{model.Paladin, "PALADIN"},
		{model.Hunter, "CAZADOR"},
		{model.Worker, "TRABAJADOR"},
		{model.Pirate, "PIRATA"},
	}

	for _, a := range archetypes {
		mod := &model.ArchetypeModifiers{
			Evasion:          1.0,
			MeleeAttack:      1.0,
			ProjectileAttack: 1.0,
			WrestlingAttack:  1.0,
			MeleeDamage:      1.0,
			ProjectileDamage: 1.0,
			WrestlingDamage:  1.0,
			Shield:           1.0,
			HP:               10.0,
		}

		if sec, ok := data["MODEVASION"]; ok {
			if val, ok := sec[a.name]; ok {
				mod.Evasion = toFloat32(val)
			}
		}
		if sec, ok := data["MODATAQUEARMAS"]; ok {
			if val, ok := sec[a.name]; ok {
				mod.MeleeAttack = toFloat32(val)
			}
		}
		if sec, ok := data["MODATAQUEPROYECTILES"]; ok {
			if val, ok := sec[a.name]; ok {
				mod.ProjectileAttack = toFloat32(val)
			}
		}
		if sec, ok := data["MODDAÑOARMAS"]; ok {
			if val, ok := sec[a.name]; ok {
				mod.MeleeDamage = toFloat32(val)
			}
		}
		if sec, ok := data["MODDAÑOPROYECTILES"]; ok {
			if val, ok := sec[a.name]; ok {
				mod.ProjectileDamage = toFloat32(val)
			}
		}
		if sec, ok := data["MODDAÑOWRESTLING"]; ok {
			if val, ok := sec[a.name]; ok {
				mod.WrestlingDamage = toFloat32(val)
			}
		}
		if sec, ok := data["MODESCUDO"]; ok {
			if val, ok := sec[a.name]; ok {
				mod.Shield = toFloat32(val)
			}
		}
		if sec, ok := data["MODVIDA"]; ok {
			if val, ok := sec[a.name]; ok {
				mod.HP = toFloat32(val)
			}
		}

		archetypeModifiers[a.id] = mod
	}

	return archetypeModifiers, raceModifiers, nil
}

func toFloat32(s string) float32 {
	v, err := strconv.ParseFloat(s, 32)
	if err != nil {
		fmt.Printf("Error parsing float %s: %v\n", s, err)
		return 1.0
	}
	return float32(v)
}
