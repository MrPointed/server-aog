package model

type ObjectType int

const (
	OTFood ObjectType = iota + 1
	OTWeapon
	OTArmor
	OTTree
	OTMoney
	OTDoor
	OTContainer
	OTSign
	OTKey
	OTForum
	OTPotion
	OTBook
	OTDrink
	OTWood
	OTBonfire
	OTShield
	OTHelmet
	OTRing
	OTTeleport
	OTFurniture
	OTJewel
	OTDeposit
	OTMetal
	OTParchment
	
	OTBoat ObjectType = 31
)

type Object struct {
	ID           int
	Name         string
	GraphicIndex int
	Type         ObjectType
	Value        int
	
	// Stats
	MinHit int
	MaxHit int
	MinDef int
	MaxDef int
	
	// Requirements
	MinLevel int
	Newbie   bool
	NoDrop   bool
	OnlyMen  bool
	OnlyWomen bool
	DwarfOnly bool
	DarkElfOnly bool
	OnlyRoyal bool
	OnlyChaos bool
	
	// Consumables
	HungerPoints int
	ThirstPoints int
	
	// Potions
	PotionType  int
	MaxModifier int
	MinModifier int
	Duration    int
	
	// Requirements / Restrictions
	ForbiddenArchetypes []UserArchetype
	
	// Graphics
	EquippedWeaponGraphic int
	EquippedArmorGraphic  int
	EquippedHelmetGraphic int
	
	// Map interaction
	Pickupable bool
	Ranged     bool

	// Doors
	OpenIndex   int
	ClosedIndex int
}

type WorldObject struct {
	Object *Object
	Amount int
}
