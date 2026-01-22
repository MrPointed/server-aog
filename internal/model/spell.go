package model

type SpellTarget int

const (
	TargetUser SpellTarget = iota + 1
	TargetNpc
	TargetUserAndNpc
	TargetTerrain
)

type SpellType int

const (
	STRevive SpellType = iota + 1
	STCurse      // Poison
	STBlessing   // Buff
	STAttack     // Damage
	STSummon
	STTeleport
)

type Spell struct {
	ID             int
	Name           string
	Description    string
	MagicWords     string
	CasterMsg      string
	OwnMsg         string
	TargetMsg      string
	Type           int // Using raw for now until mappings are clear
	WAV            int
	FX             int
	Loops          int
	MinSkill       int
	ManaRequired   int
	StaminaRequired int
	TargetType     SpellTarget
	
	// Stats
	MinHP int
	MaxHP int
	
	Invisibility bool
	Paralyzes    bool
	Immobilizes  bool
}
