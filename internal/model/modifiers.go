package model

type ArchetypeModifiers struct {
	Evasion           float32
	MeleeAttack       float32
	ProjectileAttack  float32
	WrestlingAttack   float32
	MeleeDamage       float32
	ProjectileDamage  float32
	WrestlingDamage   float32
	Shield            float32
	HP                float32
}

type RaceModifiers struct {
	Strength     int
	Dexterity    int
	Intelligence int
	Charisma     int
	Constitution int
}

type GlobalBalanceConfig struct {
	EnteraDist      []int
	SemienteraDist  []int
	LevelExponent   float64
	ManaRecoveryPct int
}