package model

type ArchetypeModifiers struct {
	Evasion          float32
	MeleeAttack      float32
	ProjectileAttack float32
	WrestlingAttack  float32
	MeleeDamage      float32
	ProjectileDamage float32
	WrestlingDamage  float32
	Shield           float32
	HP               float32
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
	ManaRecoveryPct float64
	NwMaxLevel      int

	// Intervals in milliseconds
	IntervalAttack          int64
	IntervalSpell           int64
	IntervalItem            int64
	IntervalWork            int64
	IntervalMagicHit        int64
	IntervalHunger          int64
	IntervalThirst          int64
	IntervalStartMeditating int64
	IntervalMeditation      int64
	IntervalParalyzed       int64

	// NPC Intervals in milliseconds
	NPCIntervalMove   int64
	NPCIntervalAttack int64
	NPCParalizedTime  int64
}
