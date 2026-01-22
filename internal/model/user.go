package model

import "time"

type Account struct {
	Nick       string
	Password   string
	Mail       string
	Characters []string
	Banned     bool
}

func (a *Account) Authenticate(password string) bool {
	// Simple comparison, should be hashed in real scenario
	return a.Password == password
}

func (a *Account) HasCharacter(name string) bool {
	for _, c := range a.Characters {
		if c == name {
			return true
		}
	}
	return false
}

type Character struct {
	Name        string
	Description string
	Race        Race
	Gender      Gender
	Archetype   UserArchetype
	Level       byte
	Exp         int

	Hp          int
	MaxHp       int
	Mana        int
	MaxMana     int
	Stamina     int
	MaxStamina  int
	Hunger      int
	Thirstiness int

	Attributes map[Attribute]byte
	Skills     map[Skill]int

	Position Position
	Heading  Heading

	Body         int
	Head         int
	OriginalHead int

	Weapon int16
	Shield int16
	Helmet int16

	CharIndex int16

	Inventory Inventory
	Spells    []int
	SelectedSpell int

	Poisoned    bool
	Paralyzed   bool
	Immobilized bool
	Invisible   bool
	Meditating  bool
	Sailing     bool
	Dead        bool
	Hidden      bool

	// Targets
	TargetMap     int
	TargetX       int
	TargetY       int
	TargetObj     int
	TargetObjMap  int
	TargetObjX    int
	TargetObjY    int
	TargetUser    int16
	TargetNPC     int16
	TargetNpcType NPCType

	// Action Timestamps
	LastAttack time.Time
	LastSpell  time.Time
	LastItem   time.Time
	LastWork   time.Time
}

type InventorySlot struct {
	ObjectID int
	Amount   int
	Equipped bool
}

const InventorySlots = 30

type Inventory struct {
	Slots [InventorySlots]InventorySlot
}

func (inv *Inventory) GetSlot(idx int) *InventorySlot {
	if idx < 0 || idx >= InventorySlots {
		return nil
	}
	return &inv.Slots[idx]
}

func (inv *Inventory) FindEmptySlot() int {
	for i := 0; i < InventorySlots; i++ {
		if inv.Slots[i].ObjectID == 0 {
			return i
		}
	}
	return -1
}

func (inv *Inventory) AddItem(objectID int, amount int) bool {
	// Try to stack if not equipment (simplified: stack everything for now)
	for i := 0; i < InventorySlots; i++ {
		if inv.Slots[i].ObjectID == objectID {
			inv.Slots[i].Amount += amount
			return true
		}
	}

	slot := inv.FindEmptySlot()
	if slot != -1 {
		inv.Slots[slot] = InventorySlot{ObjectID: objectID, Amount: amount}
		return true
	}
	return false
}

func NewCharacter(name string, race Race, gender Gender, archetype UserArchetype) *Character {
	return &Character{
		Name:      name,
		Race:      race,
		Gender:    gender,
		Archetype: archetype,
		Level:     1,
		Attributes: make(map[Attribute]byte),
		Skills:     make(map[Skill]int),
	}
}
