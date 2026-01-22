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
	
	// Consumables
	HungerPoints int
	ThirstPoints int
	
	// Requirements / Restrictions
	ForbiddenArchetypes []UserArchetype
	
	// Graphics
	EquippedWeaponGraphic int
	EquippedBodyGraphic   int
	EquippedHeadGraphic   int
	
	// Map interaction
	Pickupable bool
}

type WorldObject struct {
	Object *Object
	Amount int
}
