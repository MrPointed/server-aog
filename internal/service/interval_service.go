package service

import (
	"time"
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/config"
)

type IntervalService struct {
	config *config.Config
}

func NewIntervalService(cfg *config.Config) *IntervalService {
	return &IntervalService{
		config: cfg,
	}
}

func (s *IntervalService) CanAttack(char *model.Character) bool {
	now := time.Now()
	
	// Check standard attack interval
	if now.Sub(char.LastAttack).Milliseconds() < s.config.IntervalAttack {
		return false
	}

	// Check Magic-Hit interval (If they casted a spell recently, they might have to wait)
	if now.Sub(char.LastSpell).Milliseconds() < s.config.IntervalMagicHit {
		return false
	}

	return true
}

func (s *IntervalService) CanCastSpell(char *model.Character) bool {
	now := time.Now()

	// Check standard spell interval
	if now.Sub(char.LastSpell).Milliseconds() < s.config.IntervalSpell {
		return false
	}

	// Check Hit-Magic interval (If they attacked recently, they might have to wait)
	if now.Sub(char.LastAttack).Milliseconds() < s.config.IntervalMagicHit {
		return false
	}

	return true
}

func (s *IntervalService) CanUseItem(char *model.Character) bool {
	now := time.Now()
	if now.Sub(char.LastItem).Milliseconds() < s.config.IntervalItem {
		return false
	}
	return true
}

func (s *IntervalService) CanWork(char *model.Character) bool {
	now := time.Now()
	if now.Sub(char.LastWork).Milliseconds() < s.config.IntervalWork {
		return false
	}
	return true
}

func (s *IntervalService) CanNPCAttack(npc *model.WorldNPC) bool {
	now := time.Now()

	// Check standard attack interval
	if now.Sub(npc.LastAttack).Milliseconds() < s.config.IntervalAttack {
		return false
	}

	// Check Magic-Hit interval
	if now.Sub(npc.LastSpell).Milliseconds() < s.config.IntervalMagicHit {
		return false
	}

	return true
}

func (s *IntervalService) CanNPCCastSpell(npc *model.WorldNPC) bool {
	now := time.Now()

	// Check standard spell interval
	if now.Sub(npc.LastSpell).Milliseconds() < s.config.IntervalSpell {
		return false
	}

	// Check Hit-Magic interval
	if now.Sub(npc.LastAttack).Milliseconds() < s.config.IntervalMagicHit {
		return false
	}

	return true
}

func (s *IntervalService) UpdateLastAttack(char *model.Character) {
	char.LastAttack = time.Now()
}

func (s *IntervalService) UpdateNPCLastAttack(npc *model.WorldNPC) {
	npc.LastAttack = time.Now()
}

func (s *IntervalService) UpdateNPCLastSpell(npc *model.WorldNPC) {
	npc.LastSpell = time.Now()
}

func (s *IntervalService) UpdateLastSpell(char *model.Character) {
	char.LastSpell = time.Now()
}

func (s *IntervalService) UpdateLastItem(char *model.Character) {
	char.LastItem = time.Now()
}

func (s *IntervalService) UpdateLastWork(char *model.Character) {
	char.LastWork = time.Now()
}
