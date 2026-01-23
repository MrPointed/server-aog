package service

import (
	"time"

	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/utils"
)

type TimedEventsService struct {
	userService    *UserService
	messageService *MessageService
	stopChan       chan struct{}
}

func NewTimedEventsService(userService *UserService, messageService *MessageService) *TimedEventsService {
	return &TimedEventsService{
		userService:    userService,
		messageService: messageService,
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

		// HP Regen (base on Constitution)
		if char.Hp < char.MaxHp {
			regen := int(char.Attributes[model.Constitution] / 5)
			if regen < 1 { regen = 1 }
			char.Hp = utils.Min(char.MaxHp, char.Hp+regen)
			changed = true
		}

		// Mana Regen (base on Intelligence)
		if char.Mana < char.MaxMana {
			regen := int(char.Attributes[model.Intelligence] / 3)
			if regen < 1 { regen = 1 }
			
			if char.Meditating {
				regen *= 3
			}
			
			char.Mana = utils.Min(char.MaxMana, char.Mana+regen)
			changed = true
		}

		// Stamina Regen
		if char.Stamina < char.MaxStamina {
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

		// Hunger and Thirst
		char.Hunger = utils.Min(100, char.Hunger+1)
		char.Thirstiness = utils.Min(100, char.Thirstiness+1)
		
		if char.Hunger >= 100 || char.Thirstiness >= 100 {
			char.Hp -= 1
			if char.Hp <= 0 {
				s.messageService.HandleDeath(char, "Has muerto de hambre o sed.")
			}
			changed = true
		}

		if changed {
			conn := s.userService.GetConnection(char)
			if conn != nil {
				conn.Send(outgoing.NewUpdateUserStatsPacket(char))
				if char.Hunger%10 == 0 || char.Thirstiness%10 == 0 {
					conn.Send(&outgoing.UpdateHungerAndThirstPacket{
						MinHunger: char.Hunger, MaxHunger: 100,
						MinThirst: char.Thirstiness, MaxThirst: 100,
					})
				}
			}
		}
	}
}