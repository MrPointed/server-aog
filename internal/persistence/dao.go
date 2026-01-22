package persistence

import (
	"github.com/ao-go-server/internal/model"
)

type AccountDAO interface {
	Get(nick string) (*model.Account, error)
	Create(nick, password, mail string) (*model.Account, error)
	Exists(nick string) bool
	SaveAccount(acc *model.Account) error
}

type UserCharacterDAO interface {
	Load(nick string) (*model.Character, error)
	CreateAccountAndCharacter(nick, password, mail string, race model.Race, gender model.Gender, 
		archetype model.UserArchetype, head int, city model.City, attributes map[model.Attribute]byte) (*model.Account, *model.Character, error)
	Exists(nick string) bool
	SaveCharacter(char *model.Character) error
}
