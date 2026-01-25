package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version                  string
	CharacterCreationEnabled bool
	RestrictedToAdmins       bool
	MaxConcurrentUsers       int
	XpMultiplier             float64
	GoldMultiplier           float64

	Gods        []string
	SemiGods    []string
	Counselors  []string
	RoleMasters []string

	// Security
	MD5Enabled      bool
	AcceptedMD5s    []string
	CheckCriticalFiles bool
}

type yamlConfig struct {
	Server struct {
		Init struct {
			Version                  string `yaml:"version"`
			MaxUsers                 int    `yaml:"max_users"`
			AllowCharacterCreation bool   `yaml:"allow_character_creation"`
			OnlyGMs                  bool   `yaml:"only_gms"`
			XpMultiplier             float64 `yaml:"xp_multiplier"`
			GoldMultiplier           float64 `yaml:"gold_multiplier"`
		} `yaml:"init"`
		Gods        []string `yaml:"gods"`
		SemiGods    []string `yaml:"semi_gods"`
		Counselors  []string `yaml:"counselors"`
		RoleMasters []string `yaml:"role_masters"`
		Security struct {
			MD5Hush struct {
				Enabled           bool     `yaml:"enabled"`
				AcceptedMD5      []string `yaml:"accepted_md5"`
				CheckCriticalFiles bool     `yaml:"check_critical_files"`
			} `yaml:"md5_hush"`
		} `yaml:"security"`
	} `yaml:"server"`
}

func NewDefaultConfig() *Config {
	return &Config{
		Version:                  "0.13.0",
		CharacterCreationEnabled: true,
		RestrictedToAdmins:       false,
		MaxConcurrentUsers:       500,
		XpMultiplier:             1.0,
		GoldMultiplier:           1.0,
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var yc yamlConfig
	if err := yaml.Unmarshal(data, &yc); err != nil {
		return nil, err
	}

	cfg := NewDefaultConfig()
	cfg.Version = yc.Server.Init.Version
	cfg.MaxConcurrentUsers = yc.Server.Init.MaxUsers
	cfg.CharacterCreationEnabled = yc.Server.Init.AllowCharacterCreation
	cfg.RestrictedToAdmins = yc.Server.Init.OnlyGMs
	
	if yc.Server.Init.XpMultiplier > 0 {
		cfg.XpMultiplier = yc.Server.Init.XpMultiplier
	}
	if yc.Server.Init.GoldMultiplier > 0 {
		cfg.GoldMultiplier = yc.Server.Init.GoldMultiplier
	}

	cfg.Gods = yc.Server.Gods
	cfg.SemiGods = yc.Server.SemiGods
	cfg.Counselors = yc.Server.Counselors
	cfg.RoleMasters = yc.Server.RoleMasters

	cfg.MD5Enabled = yc.Server.Security.MD5Hush.Enabled
	cfg.AcceptedMD5s = yc.Server.Security.MD5Hush.AcceptedMD5
	cfg.CheckCriticalFiles = yc.Server.Security.MD5Hush.CheckCriticalFiles

	return cfg, nil
}