package service

import (
	"time"
	"github.com/ao-go-server/internal/model"
)

type IntervalServiceImpl struct {
	globalBalance *model.GlobalBalanceConfig
}

func NewIntervalServiceImpl(globalBalance *model.GlobalBalanceConfig) IntervalService {
	return &IntervalServiceImpl{
		globalBalance: globalBalance,
	}
}

func (s *IntervalServiceImpl) CanAttack(char *model.Character) bool {
	now := time.Now()
	
	// Check standard attack interval
	if now.Sub(char.LastAttack).Milliseconds() < s.globalBalance.IntervalAttack {
		return false
	}

	// Check Magic-Hit interval (If they casted a spell recently, they might have to wait)
	if now.Sub(char.LastSpell).Milliseconds() < s.globalBalance.IntervalMagicHit {
		return false
	}

	return true
}

func (s *IntervalServiceImpl) CanCastSpell(char *model.Character) bool {
	now := time.Now()

	// Check standard spell interval
	if now.Sub(char.LastSpell).Milliseconds() < s.globalBalance.IntervalSpell {
		return false
	}

	// Check Hit-Magic interval (If they attacked recently, they might have to wait)
	if now.Sub(char.LastAttack).Milliseconds() < s.globalBalance.IntervalMagicHit {
		return false
	}

	return true
}

func (s *IntervalServiceImpl) CanUseItem(char *model.Character) bool {
	now := time.Now()
	if now.Sub(char.LastItem).Milliseconds() < s.globalBalance.IntervalItem {
		return false
	}
	return true
}

func (s *IntervalServiceImpl) CanWork(char *model.Character) bool {
	now := time.Now()
	if now.Sub(char.LastWork).Milliseconds() < s.globalBalance.IntervalWork {
		return false
	}
	return true
}

func (s *IntervalServiceImpl) CanNPCAttack(npc *model.WorldNPC) bool {
	now := time.Now()

	// Check standard attack interval
	if now.Sub(npc.LastAttack).Milliseconds() < s.globalBalance.IntervalAttack {
		return false
	}

	// Check Magic-Hit interval
	if now.Sub(npc.LastSpell).Milliseconds() < s.globalBalance.IntervalMagicHit {
		return false
	}

	return true
}

func (s *IntervalServiceImpl) CanNPCCastSpell(npc *model.WorldNPC) bool {
	now := time.Now()

	// Check standard spell interval
	if now.Sub(npc.LastSpell).Milliseconds() < s.globalBalance.IntervalSpell {
		return false
	}

	// Check Hit-Magic interval
	if now.Sub(npc.LastAttack).Milliseconds() < s.globalBalance.IntervalMagicHit {
		return false
	}

	return true
}

func (s *IntervalServiceImpl) UpdateLastAttack(char *model.Character) {
	char.LastAttack = time.Now()
}

func (s *IntervalServiceImpl) UpdateNPCLastAttack(npc *model.WorldNPC) {
	npc.LastAttack = time.Now()
}

func (s *IntervalServiceImpl) UpdateNPCLastSpell(npc *model.WorldNPC) {
	npc.LastSpell = time.Now()
}

func (s *IntervalServiceImpl) UpdateLastSpell(char *model.Character) {
	char.LastSpell = time.Now()
}

func (s *IntervalServiceImpl) UpdateLastItem(char *model.Character) {
	char.LastItem = time.Now()
}

func (s *IntervalServiceImpl) UpdateLastWork(char *model.Character) {
	char.LastWork = time.Now()
}
