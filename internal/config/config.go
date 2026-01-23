package config

import (
	"strconv"
	"strings"

	"github.com/ao-go-server/internal/persistence"
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
	data, err := persistence.ReadINI(path)
	if err != nil {
		return nil, err
	}

	cfg := NewDefaultConfig()
	
	init := data["INIT"]
	if init != nil {
		if v, ok := init["VERSION"]; ok { cfg.Version = v }
		if v, ok := init["MAXUSERS"]; ok { cfg.MaxConcurrentUsers, _ = strconv.Atoi(v) }
		// ... load other simple settings if needed
	}

	cfg.Gods = loadList(data["DIOSES"], "DIOS")
	cfg.SemiGods = loadList(data["SEMIDIOSES"], "SEMIDIOS")
	cfg.Counselors = loadList(data["CONSEJEROS"], "CONSEJERO")
	cfg.RoleMasters = loadList(data["ROLESMASTERS"], "RM")

	return cfg, nil
}

func loadList(section map[string]string, prefix string) []string {
	var list []string
	if section == nil {
		return list
	}
	
	// Scan blindly for prefix + number? Or just iterate map since keys are DIOS1, DIOS2...
	// Iterating map is easier but order is random (doesn't matter for containment check)
	for k, v := range section {
		if strings.HasPrefix(k, prefix) {
			list = append(list, strings.ToUpper(v))
		}
	}
	return list
}
