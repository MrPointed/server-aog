package persistence

import "github.com/ao-go-server/internal/model"

type BalanceRepository interface {
	Load() (map[model.UserArchetype]*model.ArchetypeModifiers, map[model.Race]*model.RaceModifiers, *model.GlobalBalanceConfig, error)
}

type CityRepository interface {
	Load() (map[int]model.City, error)
}

type MapRepository interface {
	GetMapsAmount() int
	LoadProperties(path string) error
	Load() ([]*model.Map, error)
	LoadMap(id int) (*model.Map, error)
}

type NpcRepository interface {
	Load() (map[int]*model.NPC, error)
}

type ObjectRepository interface {
	Load() (map[int]*model.Object, error)
}

type SpellRepository interface {
	Load() (map[int]*model.Spell, error)
}

type UserRepository interface {
	Exists(nick string) bool
	Load(nick string) (*model.Character, error)
	GetAllCharacters() ([]*model.Character, error)
	Get(nick string) (*model.Account, error)
	Create(nick, password, mail string) (*model.Account, error)
	SaveAccount(acc *model.Account) error
	SaveCharacter(char *model.Character) error
	CreateAccountAndCharacter(nick, password, mail string, race model.Race, gender model.Gender,
		archetype model.UserArchetype, head int, city model.City, attributes map[model.Attribute]byte) (*model.Account, *model.Character, error)
}
