package persistence

import (
	"os"

	"github.com/ao-go-server/internal/model"
	"gopkg.in/yaml.v3"
)

type BalanceYamlRepo struct {
	path string
}

func NewBalanceYamlRepo(path string) *BalanceYamlRepo {
	return &BalanceYamlRepo{path: path}
}

type yamlBalance struct {
	Balance struct {
		Races        map[string]map[string]int     `yaml:"races"`
		Classes      map[string]map[string]float32 `yaml:"classes"`
		Distribution struct {
			E []int `yaml:"e"`
			S []int `yaml:"s"`
		} `yaml:"distribution"`
		Intervals struct {
			UserAttack      int64 `yaml:"user_attack"`
			CastSpell       int64 `yaml:"cast_spell"`
			UserUse         int64 `yaml:"user_use"`
			Work            int64 `yaml:"work"`
			MagicHit        int64 `yaml:"magic_hit"`
			Hunger          int64 `yaml:"hunger"`
			Thirst          int64 `yaml:"thirst"`
			StartMeditating int64 `yaml:"start_meditating"`
			Meditation      int64 `yaml:"meditation"`
		} `yaml:"intervals"`
		Party struct {
			LevelExponent float64 `yaml:"level_exponent"`
		} `yaml:"party"`
		Extra struct {
			ManaRecoveryPercentage float64 `yaml:"mana_recovery_percentage"`
			NewbieMaxLevel         int     `yaml:"newbie_max_level"`
		} `yaml:"extra"`
		NPC struct {
			Intervals struct {
				MoveSpeed int64 `yaml:"move_speed"`
				Attack    int64 `yaml:"attack"`
				Paralized int64 `yaml:"paralized"`
			} `yaml:"intervals"`
		} `yaml:"npc"`
	} `yaml:"balance"`
}

func (d *BalanceYamlRepo) Load() (map[model.UserArchetype]*model.ArchetypeModifiers, map[model.Race]*model.RaceModifiers, *model.GlobalBalanceConfig, error) {
	data, err := os.ReadFile(d.path)
	if err != nil {
		return nil, nil, nil, err
	}

	var yb yamlBalance
	if err := yaml.Unmarshal(data, &yb); err != nil {
		return nil, nil, nil, err
	}

	archetypeModifiers := make(map[model.UserArchetype]*model.ArchetypeModifiers)
	raceModifiers := make(map[model.Race]*model.RaceModifiers)

	globalConfig := &model.GlobalBalanceConfig{
		EnteraDist:              yb.Balance.Distribution.E,
		SemienteraDist:          yb.Balance.Distribution.S,
		LevelExponent:           yb.Balance.Party.LevelExponent,
		NwMaxLevel:              yb.Balance.Extra.NewbieMaxLevel,
		ManaRecoveryPct:         yb.Balance.Extra.ManaRecoveryPercentage,
		IntervalAttack:          yb.Balance.Intervals.UserAttack,
		IntervalSpell:           yb.Balance.Intervals.CastSpell,
		IntervalItem:            yb.Balance.Intervals.UserUse,
		IntervalWork:            yb.Balance.Intervals.Work,
		IntervalMagicHit:        yb.Balance.Intervals.MagicHit,
		IntervalHunger:          yb.Balance.Intervals.Hunger,
		IntervalThirst:          yb.Balance.Intervals.Thirst,
		IntervalStartMeditating: yb.Balance.Intervals.StartMeditating,
		IntervalMeditation:      yb.Balance.Intervals.Meditation,
		NPCIntervalMove:         yb.Balance.NPC.Intervals.MoveSpeed,
		NPCIntervalAttack:       yb.Balance.NPC.Intervals.Attack,
		NPCParalizedTime:        yb.Balance.NPC.Intervals.Paralized * 60 * 1000, // min to ms
	}

	// ... (Races and Classes mapping)
	raceMap := map[string]model.Race{
		"human": model.Human,
		"elf":   model.Elf,
		"drow":  model.DarkElf,
		"gnome": model.Gnome,
		"dwarf": model.Dwarf,
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

	return archetypeModifiers, raceModifiers, globalConfig, nil
}
