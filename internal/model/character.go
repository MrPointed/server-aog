package model

type Attribute int

const (
	Strength Attribute = iota
	Dexterity
	Intelligence
	Charisma
	Constitution
)

type Race int

const (
	Human Race = iota + 1
	Elf
	DarkElf
	Gnome
	Dwarf
)

type Gender int

const (
	Male Gender = iota + 1
	Female
)

type UserArchetype int

const (
	Mage UserArchetype = iota + 1
	Cleric
	Warrior
	Assasin
	Thief
	Bard
	Druid
	Bandit
	Paladin
	Hunter
	Worker
	Pirate
)

type Skill int

const (
	Magic Skill = iota + 1
	Steal
	Evasion
	MeleeCombat
	Meditate
	Stab
	Hiding
	Survive
	Lumber
	Trade
	Defense
	Fishing
	Mining
	Woodwork
	Ironwork
	Leadership
	Tame
	Projectiles
	Wrestling
	Sailing
)
