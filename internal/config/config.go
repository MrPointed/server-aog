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

	// Intervals in milliseconds
	IntervalAttack   int64
	IntervalSpell    int64
	IntervalItem     int64
	IntervalWork     int64
	IntervalMagicHit int64

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
		} `yaml:"init"`
		Gods        []string `yaml:"gods"`
		SemiGods    []string `yaml:"semi_gods"`
		Counselors  []string `yaml:"counselors"`
		RoleMasters []string `yaml:"role_masters"`
		Intervals   struct {
			UserAttack int64 `yaml:"user_attack"`
			CastSpell  int64 `yaml:"cast_spell"`
			UserUse    int64 `yaml:"user_use"`
			Work       int64 `yaml:"work"`
			MagicHit   int64 `yaml:"magic_hit"`
		} `yaml:"intervals"`
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

		IntervalAttack:   1500,
		IntervalSpell:    1400,
		IntervalItem:     450,
		IntervalWork:     800,
		IntervalMagicHit: 1000,
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

	cfg.IntervalAttack = yc.Server.Intervals.UserAttack
	cfg.IntervalSpell = yc.Server.Intervals.CastSpell
	cfg.IntervalItem = yc.Server.Intervals.UserUse
	cfg.IntervalWork = yc.Server.Intervals.Work
	cfg.IntervalMagicHit = yc.Server.Intervals.MagicHit

	cfg.Gods = yc.Server.Gods
	cfg.SemiGods = yc.Server.SemiGods
	cfg.Counselors = yc.Server.Counselors
	cfg.RoleMasters = yc.Server.RoleMasters

	cfg.MD5Enabled = yc.Server.Security.MD5Hush.Enabled
	cfg.AcceptedMD5s = yc.Server.Security.MD5Hush.AcceptedMD5
	cfg.CheckCriticalFiles = yc.Server.Security.MD5Hush.CheckCriticalFiles

	return cfg, nil
}
