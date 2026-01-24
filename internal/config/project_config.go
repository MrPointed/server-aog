package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

type ProjectConfig struct {
	Project struct {
		Paths struct {
			Charfiles   string `yaml:"charfiles"`
			CitiesDat   string `yaml:"cities_dat"`
			NpcsDat     string `yaml:"npcs_dat"`
			ObjectsDat  string `yaml:"objects_dat"`
			Maps        string `yaml:"maps"`
		} `yaml:"paths"`
		MapsCount int `yaml:"maps_count"`
	} `yaml:"project"`
}

func LoadProjectConfig(path string) (*ProjectConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pc ProjectConfig
	if err := yaml.Unmarshal(data, &pc); err != nil {
		return nil, err
	}

	return &pc, nil
}
