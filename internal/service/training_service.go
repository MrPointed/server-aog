package service

import (
	"fmt"
	"math"

	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/utils"
)

const (
	MaxLevel            = 50
	MaxHitUnder36       = 99
	MaxHitOver36        = 99
	MaxHP               = 999
	MaxMana             = 9999
	MaxStamina          = 9999
	DefaultStaminaGain  = 15
	ThiefStaminaGain    = 20
	MageStaminaGain     = 2
	WorkerStaminaGain   = 5
	BanditStaminaGain   = 15
)

type TrainingService struct {
	messageService *MessageService
	userService    *UserService
	archetypeMods  map[model.UserArchetype]*model.ArchetypeModifiers
}

func NewTrainingService(messageService *MessageService, userService *UserService, archetypeMods map[model.UserArchetype]*model.ArchetypeModifiers) *TrainingService {
	return &TrainingService{
		messageService: messageService,
		userService:    userService,
		archetypeMods:  archetypeMods,
	}
}

func (s *TrainingService) CheckLevel(char *model.Character) {
	if char.Level >= MaxLevel {
		char.Exp = 0
		char.ExpToNext = 0
		return
	}

	leveledUp := false
	ptsAwarded := 0

	for char.Exp >= char.ExpToNext && char.Level < MaxLevel {
		leveledUp = true
		char.Exp -= char.ExpToNext
		char.Level++

		// Award skill points
		if char.Level == 2 {
			ptsAwarded += 10
		} else {
			ptsAwarded += 5
		}

		// Calculate new EXP threshold
		s.updateExpThreshold(char)

		// Calculate gains
		hpGain := s.calculateHPGain(char)
		manaGain := s.calculateManaGain(char)
		stamGain := s.calculateStaminaGain(char)
		hitGain := s.calculateHitGain(char)

		// Apply gains
		char.MaxHp = utils.Min(MaxHP, char.MaxHp+hpGain)
		char.Hp = char.MaxHp
		char.MaxMana = utils.Min(MaxMana, char.MaxMana+manaGain)
		char.MaxStamina = utils.Min(MaxStamina, char.MaxStamina+stamGain)
		
		char.MinHit = utils.Min(MaxHitOver36, char.MinHit+hitGain)
		char.MaxHit = utils.Min(MaxHitOver36, char.MaxHit+hitGain)

		// Feedback
		s.messageService.SendConsoleMessage(char, "¡Has subido de nivel!", outgoing.INFO)
		s.messageService.SendToArea(&outgoing.PlayWavePacket{
			Wave: 6, // SND_NIVEL
			X:    char.Position.X,
			Y:    char.Position.Y,
		}, char.Position)

		if hpGain > 0 {
			s.messageService.SendConsoleMessage(char, fmt.Sprintf("Has ganado %d puntos de vida.", hpGain), outgoing.INFO)
		}
		if manaGain > 0 {
			s.messageService.SendConsoleMessage(char, fmt.Sprintf("Has ganado %d puntos de mana.", manaGain), outgoing.INFO)
		}
		if stamGain > 0 {
			s.messageService.SendConsoleMessage(char, fmt.Sprintf("Has ganado %d puntos de energia.", stamGain), outgoing.INFO)
		}
		if hitGain > 0 {
			s.messageService.SendConsoleMessage(char, fmt.Sprintf("Tu golpe aumento en %d puntos.", hitGain), outgoing.INFO)
		}
	}

	if leveledUp {
		char.SkillPoints += ptsAwarded
		if ptsAwarded > 0 {
			s.messageService.SendConsoleMessage(char, fmt.Sprintf("Has ganado un total de %d skillpoints.", ptsAwarded), outgoing.INFO)
		}
		
		conn := s.userService.GetConnection(char)
		if conn != nil {
			conn.Send(outgoing.NewUpdateUserStatsPacket(char))
		}
		
		// Broadcast level up visual change if any
		s.messageService.SendToArea(&outgoing.CharacterChangePacket{Character: char}, char.Position)
	}
}

func (s *TrainingService) updateExpThreshold(char *model.Character) {
	level := int(char.Level)
	multiplier := 1.0
	
	if level < 15 {
		multiplier = 1.4
	} else if level < 21 {
		multiplier = 1.35
	} else if level < 26 {
		multiplier = 1.3
	} else if level < 35 {
		multiplier = 1.2
	} else if level < 40 {
		multiplier = 1.3
	} else {
		multiplier = 1.375
	}
	
	char.ExpToNext = int(float64(char.ExpToNext) * multiplier)
}

func (s *TrainingService) calculateHPGain(char *model.Character) int {
	mod := s.archetypeMods[char.Archetype]
	if mod == nil {
		return 0
	}

	// Promedio = ModVida(Clase) - (21 - Constitution) * 0.5
	constitution := int(char.Attributes[model.Constitution])
	promedio := float64(mod.HP) - float64(21-constitution)*0.5
	
	random := utils.RandomNumber(0, 100)
	
	// Implementation of AO HP gain distribution
	if math.Mod(promedio, 1.0) == 0.5 {
		// Semientera distribution (20, 30, 30, 20)
		if random <= 20 {
			return int(promedio + 1.5)
		} else if random <= 50 {
			return int(promedio + 0.5)
		} else if random <= 80 {
			return int(promedio - 0.5)
		} else {
			return int(promedio - 1.5)
		}
	} else {
		// Entera distribution (9, 23, 36, 23, 9)
		if random <= 9 {
			return int(promedio + 2)
		} else if random <= 32 {
			return int(promedio + 1)
		} else if random <= 68 {
			return int(promedio)
		} else if random <= 91 {
			return int(promedio - 1)
		} else {
			return int(promedio - 2)
		}
	}
}

func (s *TrainingService) calculateManaGain(char *model.Character) int {
	intelligence := int(char.Attributes[model.Intelligence])
	
	switch char.Archetype {
	case model.Paladin, model.Assasin:
		return intelligence
	case model.Mage:
		return int(2.8 * float64(intelligence))
	case model.Cleric, model.Druid, model.Bard:
		return 2 * intelligence
	case model.Bandit:
		return int(float64(intelligence) * 2.0 / 3.0)
	}
	return 0
}

func (s *TrainingService) calculateStaminaGain(char *model.Character) int {
	switch char.Archetype {
	case model.Thief:
		return ThiefStaminaGain
	case model.Mage:
		return MageStaminaGain
	case model.Worker:
		return WorkerStaminaGain
	case model.Bandit:
		return BanditStaminaGain
	default:
		return DefaultStaminaGain
	}
}

func (s *TrainingService) calculateHitGain(char *model.Character) int {
	switch char.Archetype {
	case model.Warrior, model.Hunter, model.Assasin, model.Paladin:
		if char.Level > 35 {
			return 2
		}
		return 3
	case model.Pirate:
		return 3
	case model.Thief, model.Worker, model.Cleric, model.Druid, model.Bard, model.Bandit:
		return 2
	case model.Mage:
		return 1
	default:
		return 2
	}
}
