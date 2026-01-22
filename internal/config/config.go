package config

type Config struct {
	Version                  string
	CharacterCreationEnabled bool
	RestrictedToAdmins       bool
	MaxConcurrentUsers       int

	// Intervals in milliseconds
	IntervalAttack   int64
	IntervalSpell    int64
	IntervalItem     int64
	IntervalWork     int64
	IntervalMagicHit int64
}

func NewDefaultConfig() *Config {
	return &Config{
		Version:                  "0.13.0",
		CharacterCreationEnabled: true,
		RestrictedToAdmins:       false,
		MaxConcurrentUsers:       500,

		IntervalAttack:   1500,
		IntervalSpell:    1400,
		IntervalItem:     450,
		IntervalWork:     800,
		IntervalMagicHit: 1000,
	}
}
