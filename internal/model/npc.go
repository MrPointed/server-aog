package model

import "time"

type NPCType int

const (
	NTCommon       NPCType = iota // 0
	NTHealer                      // 1
	NTGuard                       // 2
	NTTrainer                     // 3
	NTBanker                      // 4
	NTNoble                       // 5
	NTDragon                      // 6
	NTGambler                     // 7 (Timberos)
	NTGuardCaos                   // 8
	NTHealerNewbie                // 9
)

type MovementType int

const (
	MovementStatic                MovementType = 1
	MovementRandom                MovementType = 2
	MovementHostile               MovementType = 3
	MovementDefense               MovementType = 4
	MovementGuardAttackCriminals  MovementType = 5
	MovementObject                MovementType = 6
	MovementFollowOwner           MovementType = 8
	MovementAttackNpc             MovementType = 9
	MovementPathfinding           MovementType = 10
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

	AttackPower  int
	EvasionPower int
	Defense      int
	MagicDefense int

	Hostile bool

	// Template AI properties
	LanzaSpells int
	Spells      []int
	AtacaDoble  bool
	ReSpawn     bool

	// Trading
	CanTrade  bool
	Inventory []InventorySlot

	// Spawning
	Movement int

	Drops []NPCDrop
}

type NPCDrop struct {
	ObjectID int
	Amount   int
}

type WorldNPC struct {
	NPC          *NPC
	Position     Position
	Heading      Heading
	HP           int
	RemainingExp int
	Index        int16

	// Stateful AI flags
	Inmovilizado bool
	Paralizado   bool
	ParalizadoSince time.Time
	OldMovement  int
	OldHostile   bool
	AttackedBy   string
	Follow       bool
	MaestroUser  int // Index of the user who owns this NPC
	Respawn      bool

	// Intervals
	LastAttack time.Time
	LastSpell  time.Time
	LastMovement time.Time
}
