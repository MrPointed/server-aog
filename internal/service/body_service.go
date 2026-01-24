package service

import (
	"strconv"
	"strings"

	"github.com/ao-go-server/internal/config"
	"github.com/ao-go-server/internal/model"
)

type CharacterBodyService struct {
	// Simple maps for validation
	validHeads    map[model.Race]map[model.Gender][]int
	defaultBodies map[model.Race]map[model.Gender]int
}

func NewCharacterBodyService(cfg *config.ProjectConfig) *CharacterBodyService {
	s := &CharacterBodyService{
		validHeads:    make(map[model.Race]map[model.Gender][]int),
		defaultBodies: make(map[model.Race]map[model.Gender]int),
	}
	s.setupFromConfig(cfg)
	return s
}

func (s *CharacterBodyService) setupFromConfig(cfg *config.ProjectConfig) {
	raceMap := map[string]model.Race{
		"human":   model.Human,
		"elf":     model.Elf,
		"darkelf": model.DarkElf,
		"gnome":   model.Gnome,
		"dwarf":   model.Dwarf,
	}

	for name, r := range raceMap {
		s.validHeads[r] = make(map[model.Gender][]int)
		s.defaultBodies[r] = make(map[model.Gender]int)

		if raceCfg, ok := cfg.Project.Races[name]; ok {
			// Heads
			s.validHeads[r][model.Male] = s.parseHeadRange(raceCfg.Heads.Male)
			s.validHeads[r][model.Female] = s.parseHeadRange(raceCfg.Heads.Female)

			// Bodies
			s.defaultBodies[r][model.Male] = raceCfg.Bodies.Male
			s.defaultBodies[r][model.Female] = raceCfg.Bodies.Female
		}
	}
}

func (s *CharacterBodyService) parseHeadRange(r string) []int {
	var heads []int
	parts := strings.Split(r, "-")
	if len(parts) == 2 {
		start, _ := strconv.Atoi(parts[0])
		end, _ := strconv.Atoi(parts[1])
		for i := start; i <= end; i++ {
			heads = append(heads, i)
		}
	} else if val, err := strconv.Atoi(r); err == nil {
		heads = append(heads, val)
	}
	return heads
}

func (s *CharacterBodyService) IsValidHead(head int, race model.Race, gender model.Gender) bool {
	heads, ok := s.validHeads[race][gender]
	if !ok {
		return false
	}
	for _, h := range heads {
		if h == head {
			return true
		}
	}
	return false
}

func (s *CharacterBodyService) GetBody(race model.Race, gender model.Gender) int {
	return s.defaultBodies[race][gender]
}
