package service

import (
	"log/slog"
	"time"

	"github.com/ao-go-server/internal/config"
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/utils"
)

type TimedEventsServiceImpl struct {
	userService    UserService
	messageService MessageService
	loginService   LoginService
	config         *config.Config
	globalBalance  *model.GlobalBalanceConfig
	stopChan       chan struct{}
}

func NewTimedEventsServiceImpl(userService UserService, messageService MessageService, loginService LoginService, cfg *config.Config, globalBalance *model.GlobalBalanceConfig) TimedEventsService {
	return &TimedEventsServiceImpl{
		userService:    userService,
		messageService: messageService,
		loginService:   loginService,
		config:         cfg,
		globalBalance:  globalBalance,
		stopChan:       make(chan struct{}),
	}
}

func (s *TimedEventsServiceImpl) Start() {
	go s.regenLoop()
	go s.worldSaveLoop()
}

func (s *TimedEventsServiceImpl) Stop() {
	close(s.stopChan)
}

func (s *TimedEventsServiceImpl) regenLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.processRegen()
		case <-s.stopChan:
			return
		}
	}
}

func (s *TimedEventsServiceImpl) worldSaveLoop() {
	// World Save interval from config (in minutes)
	if s.config.WorldSaveInterval <= 0 {
		slog.Info("Automatic WorldSave is disabled (interval set to 0)")
		return
	}

	interval := time.Duration(s.config.WorldSaveInterval) * time.Minute
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.loginService.WorldSave()
		case <-s.stopChan:
			return
		}
	}
}

func (s *TimedEventsServiceImpl) processRegen() {
	chars := s.userService.GetLoggedCharacters()
	now := time.Now()

	for _, char := range chars {
		if char.Dead {
			continue
		}

		changed := false
		canRegen := char.Hunger > 0 && char.Thirstiness > 0

		// HP Regen (base on Constitution) - Every 2 seconds
		if canRegen && char.Hp < char.MaxHp && now.Sub(char.LastHPRegen).Seconds() >= 2 {
			regen := int(char.Attributes[model.Constitution] / 5)
			if regen < 1 {
				regen = 1
			}
			char.Hp = utils.Min(char.MaxHp, char.Hp+regen)
			char.LastHPRegen = now
			changed = true
		}

		// Mana Regen
		if canRegen && char.Mana < char.MaxMana {
			if char.Meditating {

				fxPacket := &outgoing.CreateFxPacket{
					CharIndex: char.CharIndex,
					FxID:      4,
					Loops:     -1,
				}
				if conn := s.userService.GetConnection(char); conn != nil {
					conn.Send(fxPacket)
				}
				s.messageService.AreaService().BroadcastNearby(char, fxPacket)
				// Check if meditation start delay has passed
				if now.Sub(char.MeditatingSince).Milliseconds() >= s.globalBalance.IntervalStartMeditating {
					// Check meditation interval
					if now.Sub(char.LastMeditationRegen).Milliseconds() >= s.globalBalance.IntervalMeditation {
						regen := int(float64(char.MaxMana+char.Skills[model.Meditate]) * s.globalBalance.ManaRecoveryPct)
						char.Mana = utils.Min(char.MaxMana, char.Mana+regen)
						char.LastMeditationRegen = now
						changed = true
					}
				}
			} else if now.Sub(char.LastManaRegen).Seconds() >= 2 {
				// Base Mana Regen (base on Intelligence) - Every 2 seconds
				regen := int(char.Attributes[model.Intelligence] / 3)
				if regen < 1 {
					regen = 1
				}
				char.Mana = utils.Min(char.MaxMana, char.Mana+regen)
				char.LastManaRegen = now
				changed = true
			}
		}

		// Stamina Regen - Every 2 seconds
		if canRegen && char.Stamina < char.MaxStamina && now.Sub(char.LastStaminaRegen).Seconds() >= 2 {
			regen := 5
			char.Stamina = utils.Min(char.MaxStamina, char.Stamina+regen)
			char.LastStaminaRegen = now
			changed = true
		}

		// Poison damage - Every 2 seconds (using LastHPRegen as a proxy or just hardcoded for now, ideally separate)
		if char.Poisoned && now.Sub(char.LastHPRegen).Seconds() >= 2 {
			// Actually HP regen and poison should probably have their own intervals.
			// For simplicity let's assume they were tied to the 2s ticker.
			damage := utils.RandomNumber(1, 5)
			char.Hp -= damage
			if char.Hp <= 0 {
				s.messageService.HandleDeath(char, "Has muerto por el veneno.")
			}
			changed = true
		}

		// Hunger and Thirst
		if char.LastHungerUpdate.IsZero() {
			char.LastHungerUpdate = now
		}
		if char.LastThirstUpdate.IsZero() {
			char.LastThirstUpdate = now
		}

		if now.Sub(char.LastHungerUpdate).Milliseconds() >= s.globalBalance.IntervalHunger {
			if char.Hunger > 0 {
				decay := utils.RandomNumber(1, 3)
				char.Hunger = utils.Max(0, char.Hunger-decay)
				changed = true
			}
			char.LastHungerUpdate = now
		}

		if now.Sub(char.LastThirstUpdate).Milliseconds() >= s.globalBalance.IntervalThirst {
			if char.Thirstiness > 0 {
				decay := utils.RandomNumber(1, 3)
				char.Thirstiness = utils.Max(0, char.Thirstiness-decay)
				changed = true
			}
			char.LastThirstUpdate = now
		}

		if char.Hunger <= 0 || char.Thirstiness <= 0 {
			// HP decay from hunger/thirst - Every 2 seconds
			if now.Sub(char.LastHPRegen).Seconds() >= 2 {
				char.Hp -= 1
				if char.Hp <= 0 {
					s.messageService.HandleDeath(char, "Has muerto de hambre o sed.")
				}
				changed = true
			}
		}

		// Paralysis / Immobilize Expiration
		if (char.Paralyzed || char.Immobilized) && !char.ParalyzedSince.IsZero() {
			if now.Sub(char.ParalyzedSince).Milliseconds() >= s.globalBalance.IntervalParalyzed {
				char.Paralyzed = false
				char.Immobilized = false
				char.ParalyzedSince = time.Time{}
				s.messageService.SendConsoleMessage(char, "¡Has recuperado el movimiento!", outgoing.INFO)
				if conn := s.userService.GetConnection(char); conn != nil {
					conn.Send(&outgoing.ParalyzeOkPacket{})
				}
			}
		}

		// Potion Effects Expiration
		if !char.StrengthEffectEnd.IsZero() && now.After(char.StrengthEffectEnd) {
			char.Attributes[model.Strength] = char.OriginalAttributes[model.Strength]
			char.StrengthEffectEnd = time.Time{}
			s.messageService.SendConsoleMessage(char, "El efecto de la poción de fuerza ha terminado.", outgoing.INFO)
			conn := s.userService.GetConnection(char)
			if conn != nil {
				conn.Send(&outgoing.UpdateStrengthAndDexterityPacket{
					Strength:  char.Attributes[model.Strength],
					Dexterity: char.Attributes[model.Dexterity],
				})
			}
		}
		if !char.AgilityEffectEnd.IsZero() && now.After(char.AgilityEffectEnd) {
			char.Attributes[model.Dexterity] = char.OriginalAttributes[model.Dexterity]
			char.AgilityEffectEnd = time.Time{}
			s.messageService.SendConsoleMessage(char, "El efecto de la poción de agilidad ha terminado.", outgoing.INFO)
			conn := s.userService.GetConnection(char)
			if conn != nil {
				conn.Send(&outgoing.UpdateStrengthAndDexterityPacket{
					Strength:  char.Attributes[model.Strength],
					Dexterity: char.Attributes[model.Dexterity],
				})
			}
		}

		if changed {
			conn := s.userService.GetConnection(char)
			if conn != nil {
				conn.Send(outgoing.NewUpdateUserStatsPacket(char))
				conn.Send(&outgoing.UpdateHungerAndThirstPacket{
					MinHunger: char.Hunger, MaxHunger: 100,
					MinThirst: char.Thirstiness, MaxThirst: 100,
				})
			}
		}
	}
}
