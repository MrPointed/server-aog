package config

type Config struct {
	Version                  string
	CharacterCreationEnabled bool
	RestrictedToAdmins       bool
	MaxConcurrentUsers       int
}

func NewDefaultConfig() *Config {
	return &Config{
		Version:                  "0.13.0",
		CharacterCreationEnabled: true,
		RestrictedToAdmins:       false,
		MaxConcurrentUsers:       500,
	}
}
