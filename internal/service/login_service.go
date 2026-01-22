package service

import (
	"fmt"
	"github.com/ao-go-server/internal/config"
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/persistence"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/actions"
)

type LoginService struct {
	accDAO      persistence.AccountDAO
	charDAO     persistence.UserCharacterDAO
	config      *config.Config
		userService  *UserService
		mapService   *MapService
			bodyService  *CharacterBodyService
				indexManager *CharacterIndexManager
				messageService *MessageService
					objectService  *ObjectService
					cityService    *CityService
					spellService   *SpellService
					executor     *actions.ActionExecutor[*MapService]
				}
				
				func NewLoginService(accDAO persistence.AccountDAO, charDAO persistence.UserCharacterDAO, 
					cfg *config.Config, userService *UserService, mapService *MapService, 
					bodyService *CharacterBodyService, indexManager *CharacterIndexManager, 
					messageService *MessageService, objectService *ObjectService, cityService *CityService, 
					spellService *SpellService, executor *actions.ActionExecutor[*MapService]) *LoginService {
					return &LoginService{
						accDAO:       accDAO,
						charDAO:      charDAO,
						config:       cfg,
						userService:  userService,
						mapService:   mapService,
						bodyService:  bodyService,
						indexManager: indexManager,
						messageService: messageService,
						objectService:  objectService,
						cityService:    cityService,
						spellService:   spellService,
						executor:     executor,
					}
				}			
			func (s *LoginService) ConnectNewCharacter(conn protocol.Connection, nick, password, mail string, 
				raceId, genderId, archetypeId, headId, cityId byte, clientHash, version string) error {
				fmt.Printf("Connecting new character: %s\n", nick)
				
				if s.config.Version != version {
					return fmt.Errorf("versión obsoleta")
				}
			
				if !s.config.CharacterCreationEnabled {
					return fmt.Errorf("la creación de personajes está deshabilitada")
				}
			
				// Check if name taken
				if s.charDAO.Exists(nick) {
					return fmt.Errorf("el nombre ya está en uso")
				}
			
				race := model.Race(raceId)
				gender := model.Gender(genderId)
			
				if !s.bodyService.IsValidHead(int(headId), race, gender) {
					return fmt.Errorf("cabeza inválida")
				}
			
				body := s.bodyService.GetBody(race, gender)
			
				// Get attributes from dice roll
				attrs := make(map[model.Attribute]byte)
				attrs[model.Strength] = conn.GetAttribute(int(model.Strength))
				attrs[model.Dexterity] = conn.GetAttribute(int(model.Dexterity))
				attrs[model.Intelligence] = conn.GetAttribute(int(model.Intelligence))
				attrs[model.Charisma] = conn.GetAttribute(int(model.Charisma))
				attrs[model.Constitution] = conn.GetAttribute(int(model.Constitution))
			
				if attrs[model.Strength] == 0 {
					return fmt.Errorf("debe tirar los dados antes de crear un personaje")
				}
			
				city, ok := s.cityService.GetCity(int(cityId))
				if !ok {
					city = model.City{Map: 1, X: 50, Y: 50} // Default city placeholder
				}
			
				acc, char, err := s.charDAO.CreateAccountAndCharacter(nick, password, mail, 
					race, gender, model.UserArchetype(archetypeId), 
					int(headId), city, attrs)
	
	if err != nil {
		return err
	}

	if char.Dead {
		char.Body = 8
		char.Head = 500
	} else {
		char.Body = body
	}

	s.finalizeLogin(conn, acc, char)
	return nil
}

func (s *LoginService) ConnectExistingCharacter(conn protocol.Connection, nick, password, version, clientHash string) error {
	fmt.Printf("Connecting existing character: %s\n", nick)
	if s.config.Version != version {
		return fmt.Errorf("versión obsoleta")
	}

	acc, err := s.accDAO.Get(nick)
	if err != nil || acc == nil {
		return fmt.Errorf("personaje inexistente")
	}

	if acc.Password != password {
		return fmt.Errorf("contraseña incorrecta")
	}

	char, err := s.charDAO.Load(nick)
	if err != nil {
		return err
	}

	if char.Dead {
		char.Body = 8
		char.Head = 500
	}

	if char.Body == 0 {
		char.Body = s.bodyService.GetBody(char.Race, char.Gender)
	}

	s.finalizeLogin(conn, acc, char)
	return nil
}

func (s *LoginService) finalizeLogin(conn protocol.Connection, acc *model.Account, char *model.Character) {
	conn.Send(&outgoing.LoggedPacket{})
	char.CharIndex = s.indexManager.AssignIndex()
	conn.SetUser(char)
	s.userService.LogIn(conn)

	// Place character in world via action executor
	s.executor.Dispatch(func(m *MapService) {
		m.PutCharacterAtPos(char, char.Position)
	})

	// Handshake sequence
	gameMap := s.mapService.GetMap(char.Position.Map)
	if gameMap != nil {
		conn.Send(&outgoing.ChangeMapPacket{
			MapId:   gameMap.Id,
			Version: gameMap.Version,
		})

		// Notify nearby players about me
		s.messageService.SendToAreaButUser(&outgoing.CharacterCreatePacket{Character: char}, char.Position, char)

		// Notify me about nearby players
		for _, other := range gameMap.Characters {
			if other != char && s.messageService.AreaService.InRange(char.Position, other.Position) {
				conn.Send(&outgoing.CharacterCreatePacket{Character: other})
			}
		}
	}

	conn.Send(&outgoing.CharacterCreatePacket{Character: char})
	conn.Send(&outgoing.UserCharIndexInServerPacket{UserIndex: char.CharIndex})
	conn.Send(&outgoing.AreaChangedPacket{Position: char.Position})

	// Notify me about objects on the ground
	if gameMap != nil {
		for y := 0; y < model.MapHeight; y++ {
			for x := 0; x < model.MapWidth; x++ {
				tile := gameMap.GetTile(x, y)
				if tile.Object != nil {
					objPos := model.Position{X: byte(x), Y: byte(y), Map: char.Position.Map}
					if s.messageService.AreaService.InRange(char.Position, objPos) {
						conn.Send(&outgoing.ObjectCreatePacket{
							X:            byte(x),
							Y:            byte(y),
							GraphicIndex: int16(tile.Object.Object.GraphicIndex),
						})
					}
				}
				if tile.NPC != nil {
					if s.messageService.AreaService.InRange(char.Position, tile.NPC.Position) {
						conn.Send(&outgoing.NpcCreatePacket{Npc: tile.NPC})
					}
				}
			}
		}
	}

	// Send Inventory
	invCount := 0
	for i := 0; i < model.InventorySlots; i++ {
		slot := char.Inventory.Slots[i]
		if slot.ObjectID > 0 {
			obj := s.objectService.GetObject(slot.ObjectID)
			if obj != nil {
				invCount++
				conn.Send(&outgoing.ChangeInventorySlotPacket{
					Slot:     byte(i + 1),
					Object:   obj,
					Amount:   slot.Amount,
					Equipped: slot.Equipped,
				})
			}
		}
	}
	fmt.Printf("Sent %d inventory items to %s\n", invCount, char.Name)

	// Send Spells
	for i, spellID := range char.Spells {
		spell := s.spellService.GetSpell(spellID)
		if spell != nil {
			conn.Send(&outgoing.ChangeSpellSlotPacket{
				Slot:      byte(i + 1),
				SpellID:   int16(spellID),
				SpellName: spell.Name,
			})
		}
	}

	// Send initial state
	conn.Send(outgoing.NewUpdateUserStatsPacket(char))
	conn.Send(&outgoing.UpdateHungerAndThirstPacket{
		MinHunger: char.Hunger, MaxHunger: 100,
		MinThirst: char.Thirstiness, MaxThirst: 100,
	})
	conn.Send(&outgoing.UpdateStrengthAndDexterityPacket{
		Strength:  char.Attributes[model.Strength],
		Dexterity: char.Attributes[model.Dexterity],
	})

	fmt.Printf("User %s logged in at %+v\n", char.Name, char.Position)
}

func (s *LoginService) OnUserDisconnect(conn protocol.Connection) {
	char := conn.GetUser()
	if char != nil {
		fmt.Printf("User %s disconnected, saving...\n", char.Name)
		s.charDAO.SaveCharacter(char)

		// Broadcast removal
		s.messageService.SendToAreaButUser(&outgoing.CharacterRemovePacket{CharIndex: char.CharIndex}, char.Position, char)

		s.userService.LogOut(conn)
		// Remove from map
		s.executor.Dispatch(func(m *MapService) {
			m.RemoveCharacter(char)
		})
	}
}
