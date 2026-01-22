package model

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

	// Spawning
	Movement int

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
