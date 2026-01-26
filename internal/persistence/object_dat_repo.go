package persistence

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/ao-go-server/internal/model"
)

type ObjectDatRepo struct {
	path string
}

func NewObjectDatRepo(path string) *ObjectDatRepo {
	return &ObjectDatRepo{path: path}
}

func (d *ObjectDatRepo) Load() (map[int]*model.Object, error) {
	data, err := ReadINI(d.path)
	if err != nil {
		return nil, err
	}

	objects := make(map[int]*model.Object)

	for section, props := range data {
		if !strings.HasPrefix(section, "OBJ") {

			slog.Warn("Skipping invalid object section", "section", section)

			continue

		}

		id, err := strconv.Atoi(section[3:])
		if err != nil {

			slog.Warn("Skipping invalid object section", "section", section)

			continue
		}

		obj := &model.Object{
			ID:           id,
			Name:         props["NAME"],
			GraphicIndex: toInt(props["GRAPHIC_INDEX"]),
			Type:         model.ObjectType(toInt(props["OBJECT_TYPE"])),
			Value:        toInt(props["VALUE"]),
			Pickupable:   true,
		}
		
		        if obj.Name == "" {
		
		            slog.Warn("Object has no name in data file", "id", id)
		
		        }
		// Stats
		obj.MinHit = toInt(props["MIN_HIT"])
		obj.MaxHit = toInt(props["MAX_HIT"])
		obj.MinDef = toInt(props["MIN_DEF"])
		obj.MaxDef = toInt(props["MAX_DEF"])

		// Requirements
		obj.MinLevel = toInt(props["MIN_LEVEL"])
		if obj.MinLevel == 0 {
			obj.MinLevel = toInt(props["MINLEVEL"])
		}
		obj.Newbie = props["NEWBIE"] == "1"
		obj.NoDrop = props["NODROP"] == "1"

		// Boats never drop
		if obj.Type == model.OTBoat {
			obj.NoDrop = true
		}

		obj.OnlyMen = props["HOMBRE"] == "1"
		obj.OnlyWomen = props["MUJER"] == "1"
		obj.DwarfOnly = props["DWARF"] == "1"
		obj.DarkElfOnly = props["DARK_ELF"] == "1"
		obj.OnlyRoyal = props["REAL"] == "1"
		obj.OnlyChaos = props["CAOS"] == "1"

		// Consumables
		obj.HungerPoints = toInt(props["HUNGER_POINTS"])
		obj.ThirstPoints = toInt(props["THIRST_POINTS"])

		// Potions
		obj.PotionType = toInt(props["POTION_TYPE"])
		obj.MaxModifier = toInt(props["MAX_MODIFIER"])
		obj.MinModifier = toInt(props["MIN_MODIFIER"])
		obj.Duration = toInt(props["EFFECT_DURATION"])

		// Graphics
		obj.EquippedWeaponGraphic = toInt(props["EQUIPPED_WEAPON_GRAPHIC"])
		obj.EquippedArmorGraphic = toInt(props["EQUIPPED_ARMOR_GRAPHIC"])
		obj.EquippedHelmetGraphic = toInt(props["EQUIPPED_HEAD_GRAPHIC"])

		// Map interaction overrides
		if p, ok := props["PICKUPABLE"]; ok {
			obj.Pickupable = p == "1"
		} else {
			// Defaults based on type
			if obj.Type == model.OTTree || obj.Type == model.OTDeposit || obj.Type == model.OTBonfire {
				obj.Pickupable = false
			}
		}

		if p, ok := props["PROYECTIL"]; ok {
			obj.Ranged = p == "1"
		} else if p, ok := props["RANGED_WEAPON"]; ok {
			obj.Ranged = p == "1"
		}

		// Doors
		obj.OpenIndex = toInt(props["OPEN_INDEX"])
		obj.ClosedIndex = toInt(props["CLOSED_INDEX"])

		// Forbidden Archetypes (simplified parsing for now)
		for i := 1; i <= 10; i++ {
			archKey := fmt.Sprintf("FORBIDDEN_ARCHETYPE%d", i)
			if archName, ok := props[archKey]; ok {
				arch := parseArchetype(archName)
				if arch != 0 {
					obj.ForbiddenArchetypes = append(obj.ForbiddenArchetypes, arch)
				}
			}
		}

		objects[id] = obj
	}

	return objects, nil
}

func parseArchetype(name string) model.UserArchetype {
	switch strings.ToUpper(name) {
	case "MAGO": return model.Mage
	case "CLERIGO": return model.Cleric
	case "GUERRERO": return model.Warrior
	case "ASESINO": return model.Assasin
	case "LADRON": return model.Thief
	case "BARDO": return model.Bard
	case "DRUIDA": return model.Druid
	case "BANDIDO": return model.Bandit
	case "PALADIN": return model.Paladin
	case "CAZADOR": return model.Hunter
	case "TRABAJADOR": return model.Worker
	case "PIRATA": return model.Pirate
	}
	return 0
}
