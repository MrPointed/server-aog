package service

import (
	"time"

	"github.com/ao-go-server/internal/config"
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/utils"
)

type TimedEventsService struct {
	userService    *UserService
	messageService *MessageService
	config         *config.Config
	globalBalance  *model.GlobalBalanceConfig
	stopChan       chan struct{}
}

func NewTimedEventsService(userService *UserService, messageService *MessageService, cfg *config.Config, globalBalance *model.GlobalBalanceConfig) *TimedEventsService {
	return &TimedEventsService{
		userService:    userService,
		messageService: messageService,
		config:         cfg,
		globalBalance:  globalBalance,
		stopChan:       make(chan struct{}),
	}
}

func (s *TimedEventsService) Start() {
	go s.regenLoop()
}

func (s *TimedEventsService) Stop() {
	close(s.stopChan)
}

func (s *TimedEventsService) regenLoop() {
	ticker := time.NewTicker(2 * time.Second)
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

func (s *TimedEventsService) processRegen() {
	chars := s.userService.GetLoggedCharacters()
	for _, char := range chars {
		if char.Dead {
			continue
		}

		changed := false
		canRegen := char.Hunger > 0 && char.Thirstiness > 0

		// HP Regen (base on Constitution)
		if canRegen && char.Hp < char.MaxHp {
			regen := int(char.Attributes[model.Constitution] / 5)
			if regen < 1 {
				regen = 1
			}
			char.Hp = utils.Min(char.MaxHp, char.Hp+regen)
			changed = true
		}

		// Mana Regen (base on Intelligence)
		if canRegen && char.Mana < char.MaxMana {
			regen := int(char.Attributes[model.Intelligence] / 3)
			if regen < 1 {
				regen = 1
			}

			if char.Meditating {
				// Increasing mana being 100 points 100% of the time,
				// using mana_recovery_percentage as the multiplier.
				regen = int(100 * s.globalBalance.ManaRecoveryPct)
			}

			char.Mana = utils.Min(char.MaxMana, char.Mana+regen)
			changed = true
		}

		// Stamina Regen
		if canRegen && char.Stamina < char.MaxStamina {
			regen := 5
			char.Stamina = utils.Min(char.MaxStamina, char.Stamina+regen)
			changed = true
		}

		// Poison damage
		if char.Poisoned {
			damage := utils.RandomNumber(1, 5)
			char.Hp -= damage
			if char.Hp <= 0 {
				s.messageService.HandleDeath(char, "Has muerto por el veneno.")
			}
			changed = true
		}

		// Hunger and Thirst (100 is full, 0 is starving)
		now := time.Now()
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

		// Potion Effects Expiration
		now = time.Now()
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
