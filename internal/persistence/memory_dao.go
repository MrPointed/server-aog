package persistence

import (
	"fmt"
	"github.com/ao-go-server/internal/model"
)

type MemoryDAO struct {
	accounts   map[string]*model.Account
	characters map[string]*model.Character
}

func NewMemoryDAO() *MemoryDAO {
	return &MemoryDAO{
		accounts:   make(map[string]*model.Account),
		characters: make(map[string]*model.Character),
	}
}

func (d *MemoryDAO) Get(nick string) (*model.Account, error) {
	return d.accounts[nick], nil
}

func (d *MemoryDAO) Create(nick, password, mail string) (*model.Account, error) {
	if d.Exists(nick) {
		return nil, fmt.Errorf("account already exists")
	}
	acc := &model.Account{Nick: nick, Password: password, Mail: mail}
	d.accounts[nick] = acc
	return acc, nil
}

func (d *MemoryDAO) Exists(nick string) bool {
	_, ok := d.accounts[nick]
	return ok
}

func (d *MemoryDAO) Load(nick string) (*model.Character, error) {
	return d.characters[nick], nil
}

func (d *MemoryDAO) SaveAccount(acc *model.Account) error {
	return nil
}

func (d *MemoryDAO) SaveCharacter(char *model.Character) error {
	return nil
}

func (d *MemoryDAO) CreateAccountAndCharacter(nick, password, mail string, race model.Race, gender model.Gender, 
	archetype model.UserArchetype, head int, city model.City, attributes map[model.Attribute]byte) (*model.Account, *model.Character, error) {
	
	acc, err := d.Create(nick, password, mail)
	if err != nil {
		return nil, nil, err
	}

	char := model.NewCharacter(nick, race, gender, archetype)
	char.Head = head
	char.Position = model.Position{X: city.X, Y: city.Y, Map: city.Map}
	char.Attributes = attributes
	char.MaxHp = 20 // Default values
	char.Hp = 20
	char.MaxMana = 100
	char.Mana = 100
	char.MaxStamina = 100
	char.Stamina = 100
	char.Hunger = 100
	char.Thirstiness = 100

	// Add newbie items
	char.Inventory.AddItem(460, 1)  // Daga (Newbie)
	char.Inventory.AddItem(463, 1)  // Vestimentas Comunes (Newbie)
	char.Inventory.AddItem(461, 50) // Pocion Roja (Newbie)
	char.Inventory.AddItem(462, 50) // Pocion Verde (Newbie)
	char.Inventory.AddItem(855, 50) // Pocion Amarilla (Newbie)
	char.Inventory.AddItem(856, 50) // Pocion Azul (Newbie)
	
	d.characters[nick] = char
	acc.Characters = append(acc.Characters, nick)

	return acc, char, nil
}
