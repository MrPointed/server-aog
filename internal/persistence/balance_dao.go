package persistence

import (
	"os"

	"github.com/ao-go-server/internal/model"
	"gopkg.in/yaml.v3"
)

type BalanceDAO struct {
	path string
}

func NewBalanceDAO(path string) *BalanceDAO {
	return &BalanceDAO{path: path}
}

type yamlBalance struct {
	Balance struct {
		Races   map[string]map[string]int     `yaml:"races"`
		Classes map[string]map[string]float32 `yaml:"classes"`
	} `yaml:"balance"`
}

func (d *BalanceDAO) Load() (map[model.UserArchetype]*model.ArchetypeModifiers, map[model.Race]*model.RaceModifiers, error) {
	data, err := os.ReadFile(d.path)
	if err != nil {
		return nil, nil, err
	}

	var yb yamlBalance
	if err := yaml.Unmarshal(data, &yb); err != nil {
		return nil, nil, err
	}

	archetypeModifiers := make(map[model.UserArchetype]*model.ArchetypeModifiers)
	raceModifiers := make(map[model.Race]*model.RaceModifiers)

	// Map YAML races to model.Race
	raceMap := map[string]model.Race{
		"human":    model.Human,
		"elf":      model.Elf,
		"drow":     model.DarkElf,
		"gnome":    model.Gnome,
		"dwarf":    model.Dwarf,
	}

	for name, r := range raceMap {
		if mods, ok := yb.Balance.Races[name]; ok {
			raceModifiers[r] = &model.RaceModifiers{
				Strength:     mods["strength"],
				Dexterity:    mods["dexterity"],
				Intelligence: mods["intelligence"],
				Charisma:     mods["charisma"],
				Constitution: mods["constitution"],
			}
		}
	}

	// Map YAML classes to model.UserArchetype
	classMap := map[string]model.UserArchetype{
		"mago":       model.Mage,
		"clerigo":    model.Cleric,
		"guerrero":   model.Warrior,
		"asesino":    model.Assasin,
		"ladron":     model.Thief,
		"bardo":      model.Bard,
		"druida":     model.Druid,
		"bandido":    model.Bandit,
		"paladin":    model.Paladin,
		"cazador":    model.Hunter,
		"trabajador": model.Worker,
		"pirata":     model.Pirate,
	}

	for name, a := range classMap {
		if mods, ok := yb.Balance.Classes[name]; ok {
			archetypeModifiers[a] = &model.ArchetypeModifiers{
				Evasion:          mods["evasion"],
				MeleeAttack:      mods["weapon_attack"],
				ProjectileAttack: mods["ranged_attack"],
				MeleeDamage:      mods["weapon_damage"],
				ProjectileDamage: mods["ranged_damage"],
				WrestlingDamage:  mods["wrestling_damage"],
				Shield:           mods["shield"],
				HP:               mods["hp"],
			}
		}
	}

	return archetypeModifiers, raceModifiers, nil
}