package model

type NPCType int

const (
	NTNone NPCType = iota
	NTTrainer
	NTBanker
	NTMerchant
	NTHealer
	NTGuard
	NTGuardLegion
	NTGuardImperial
	NTHostile
	NTFriendly
)

type NPC struct {
	ID          int
	Name        string
	Description string
	Type        NPCType
	
	Head    int
	Body    int
	Heading Heading
	
	Level int
	Exp   int
	Hp    int
	MaxHp int
	
	MinHit int
	MaxHit int
	
	Hostile bool
	
	// Spawning
	Movement bool

	Drops []NPCDrop
}

type NPCDrop struct {
	ObjectID int
	Amount   int
}

type WorldNPC struct {
	NPC      *NPC
	Position Position
	Heading  Heading
	HP       int
	Index    int16
}
