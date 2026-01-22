package service

import (
	"github.com/ao-go-server/internal/model"
)

type CharacterBodyService struct {
	// Simple maps for validation
	validHeads map[model.Race]map[model.Gender][]int
	defaultBodies map[model.Race]map[model.Gender]int
}

func NewCharacterBodyService() *CharacterBodyService {
	s := &CharacterBodyService{
		validHeads: make(map[model.Race]map[model.Gender][]int),
		defaultBodies: make(map[model.Race]map[model.Gender]int),
	}
	s.setupDefaults()
	return s
}

func (s *CharacterBodyService) setupDefaults() {
	races := []model.Race{model.Human, model.Elf, model.DarkElf, model.Gnome, model.Dwarf}
	genders := []model.Gender{model.Male, model.Female}

	for _, r := range races {
		s.validHeads[r] = make(map[model.Gender][]int)
		s.defaultBodies[r] = make(map[model.Gender]int)
		for _, g := range genders {
			// Placeholder: allow heads 1-1000 for everyone for now
			heads := make([]int, 1000)
			for i := 0; i < 1000; i++ {
				heads[i] = i + 1
			}
			s.validHeads[r][g] = heads
		}
	}

	// Constants from client
	s.defaultBodies[model.Human][model.Male] = 21
	s.defaultBodies[model.Elf][model.Male] = 210
	s.defaultBodies[model.DarkElf][model.Male] = 32
	s.defaultBodies[model.Dwarf][model.Male] = 53
	s.defaultBodies[model.Gnome][model.Male] = 222

	s.defaultBodies[model.Human][model.Female] = 39
	s.defaultBodies[model.Elf][model.Female] = 259
	s.defaultBodies[model.DarkElf][model.Female] = 40
	s.defaultBodies[model.Dwarf][model.Female] = 60
	s.defaultBodies[model.Gnome][model.Female] = 260
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
