package service

import (
	"fmt"
	"github.com/ao-go-server/internal/actions"
	"github.com/ao-go-server/internal/config"
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/persistence"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type LoginService struct {
	accDAO         persistence.AccountDAO
	charDAO        persistence.UserCharacterDAO
	config         *config.Config
	userService    *UserService
	mapService     *MapService
	bodyService    *CharacterBodyService
	indexManager   *CharacterIndexManager
	messageService *MessageService
	objectService  *ObjectService
	cityService    *CityService
	spellService   *SpellService
	executor       *actions.ActionExecutor[*MapService]
}

func NewLoginService(accDAO persistence.AccountDAO, charDAO persistence.UserCharacterDAO,
	cfg *config.Config, userService *UserService, mapService *MapService,
	bodyService *CharacterBodyService, indexManager *CharacterIndexManager,
	messageService *MessageService, objectService *ObjectService, cityService *CityService,
	spellService *SpellService, executor *actions.ActionExecutor[*MapService]) *LoginService {
	return &LoginService{
		accDAO:         accDAO,
		charDAO:        charDAO,
		config:         cfg,
		userService:    userService,
		mapService:     mapService,
		bodyService:    bodyService,
		indexManager:   indexManager,
		messageService: messageService,
		objectService:  objectService,
		cityService:    cityService,
		spellService:   spellService,
		executor:       executor,
	}
}

// ConnectNewCharacter handles the creation and login of a new character.
func (s *LoginService) ConnectNewCharacter(conn protocol.Connection, nick, password, mail string,
	raceId, genderId, archetypeId, headId, cityId byte, clientHash, version string) error {
	fmt.Printf("Connecting new character: %s\n", nick)

	if err := s.validateLoginRequest(nick, version); err != nil {
		return err
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

	s.ensureValidCharacterState(char, body)

	s.finalizeLogin(conn, acc, char)
	return nil
}

// ConnectExistingCharacter handles the login of an existing character.
// It follows the protocol described in login_existing_char.txt strictly for logic.
func (s *LoginService) ConnectExistingCharacter(conn protocol.Connection, nick, password, version, clientHash string) error {
	fmt.Printf("Connecting existing character: %s\n", nick)

	// 1. Basic Validation (Version & Name)
	if err := s.validateLoginRequest(nick, version); err != nil {
		return err
	}

	// 2. Server Capacity Check
	if err := s.checkServerCapacity(); err != nil {
		return err
	}

	// 3. Authentication (Account existence & Password)
	acc, err := s.authenticate(nick, password)
	if err != nil {
		return err
	}

	// 4. Security Checks (Ban)
	if err := s.checkBan(acc); err != nil {
		return err
	}

	// 5. Concurrency Check (Already logged in)
	if err := s.checkAlreadyLoggedIn(nick); err != nil {
		return err
	}

	// 6. Load Character
	char, err := s.charDAO.Load(nick)
	if err != nil {
		return err
	}

	// 7. Ensure Character State (Dead/Body)
	defaultBody := s.bodyService.GetBody(char.Race, char.Gender)
	s.ensureValidCharacterState(char, defaultBody)

	// 8. Finalize Connection
	s.finalizeLogin(conn, acc, char)
	return nil
}

// validateLoginRequest checks version and name validity.
func (s *LoginService) validateLoginRequest(nick, version string) error {
	if s.config.Version != version {
		return fmt.Errorf("versión obsoleta, por favor actualice el juego")
	}
	if nick == "" {
		return fmt.Errorf("nombre inválido")
	}
	return nil
}

// checkServerCapacity checks if the server is full.
func (s *LoginService) checkServerCapacity() error {
	currentUsers := len(s.userService.GetLoggedCharacters())
	if currentUsers >= s.config.MaxConcurrentUsers {
		return fmt.Errorf("el servidor está lleno, intente más tarde")
	}
	return nil
}

// authenticate verifies account credentials.
func (s *LoginService) authenticate(nick, password string) (*model.Account, error) {
	acc, err := s.accDAO.Get(nick)
	if err != nil || acc == nil {
		return nil, fmt.Errorf("el personaje no existe")
	}

	// In production, use hash comparison
	if acc.Password != password {
		return nil, fmt.Errorf("contraseña incorrecta")
	}

	return acc, nil
}

// checkBan verifies if the account is banned.
func (s *LoginService) checkBan(acc *model.Account) error {
	if acc.Banned {
		return fmt.Errorf("acceso denegado: cuenta suspendida")
	}
	return nil
}

// checkAlreadyLoggedIn ensures the character is not already in the game.
func (s *LoginService) checkAlreadyLoggedIn(nick string) error {
	if s.userService.IsUserLoggedIn(nick) {
		return fmt.Errorf("el personaje ya se encuentra conectado")
	}
	return nil
}

// ensureValidCharacterState sets the correct body and head depending on dead/alive state.
func (s *LoginService) ensureValidCharacterState(char *model.Character, defaultBody int) {
	if char.Dead {
		char.Body = 8
		char.Head = 500
	} else if char.Body == 0 {
		char.Body = defaultBody
	}
}

func (s *LoginService) finalizeLogin(conn protocol.Connection, acc *model.Account, char *model.Character) {
	conn.Send(&outgoing.LoggedPacket{})
	
	// Assign Index
	char.CharIndex = s.indexManager.AssignIndex()
	
	// Bind to connection and service
	conn.SetUser(char)
	s.userService.LogIn(conn)

	// Validate Position (Prevent stuck in void)
	if s.mapService.IsInvalidPosition(char.Position) {
		city, ok := s.cityService.GetCity(1)
		if !ok {
			city = model.City{Map: 1, X: 50, Y: 50}
		}
		char.Position = model.Position{X: city.X, Y: city.Y, Map: city.Map}
	}

	// Dispatch to World (Thread-safe map modification)
	s.executor.Dispatch(func(m *MapService) {
		// TODO: Add collision check logic here similar to 'LegalPos' in VB6
		m.PutCharacterAtPos(char, char.Position)
	})

	// Send Handshake / Game State
	s.sendInitialGameState(conn, char)

	// Notify others
	s.messageService.SendToAreaButUser(&outgoing.CharacterCreatePacket{Character: char}, char.Position, char)
	s.messageService.AreaService.SendAreaState(char)

	fmt.Printf("User %s logged in at %+v\n", char.Name, char.Position)
}

func (s *LoginService) sendInitialGameState(conn protocol.Connection, char *model.Character) {
	// Map Info
	gameMap := s.mapService.GetMap(char.Position.Map)
	if gameMap != nil {
		conn.Send(&outgoing.ChangeMapPacket{
			MapId:   gameMap.Id,
			Version: gameMap.Version,
		})
	}

	// Self Character Info
	conn.Send(&outgoing.CharacterCreatePacket{Character: char})
	conn.Send(&outgoing.UserCharIndexInServerPacket{UserIndex: char.CharIndex})
	conn.Send(&outgoing.AreaChangedPacket{Position: char.Position})

	// Inventory & Spells
	s.sendInventory(conn, char)
	s.sendSpells(conn, char)

	// Stats
	conn.Send(outgoing.NewUpdateUserStatsPacket(char))
	conn.Send(&outgoing.UpdateHungerAndThirstPacket{
		MinHunger: char.Hunger, MaxHunger: 100,
		MinThirst: char.Thirstiness, MaxThirst: 100,
	})
	conn.Send(&outgoing.UpdateStrengthAndDexterityPacket{
		Strength:  char.Attributes[model.Strength],
		Dexterity: char.Attributes[model.Dexterity],
	})
}

func (s *LoginService) sendInventory(conn protocol.Connection, char *model.Character) {
	for i := 0; i < model.InventorySlots; i++ {
		slot := char.Inventory.Slots[i]
		if slot.ObjectID > 0 {
			obj := s.objectService.GetObject(slot.ObjectID)
			if obj != nil {
				conn.Send(&outgoing.ChangeInventorySlotPacket{
					Slot:     byte(i + 1),
					Object:   obj,
					Amount:   slot.Amount,
					Equipped: slot.Equipped,
				})
			}
		}
	}
}

func (s *LoginService) sendSpells(conn protocol.Connection, char *model.Character) {
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
}

func (s *LoginService) OnUserDisconnect(conn protocol.Connection) {
	char := conn.GetUser()
	if char != nil {
		fmt.Printf("User %s disconnected, saving...\n", char.Name)
		s.charDAO.SaveCharacter(char)

		// Broadcast removal
		s.messageService.SendToAreaButUser(&outgoing.CharacterRemovePacket{CharIndex: char.CharIndex}, char.Position, char)

		s.userService.LogOut(conn)
		s.indexManager.FreeIndex(char.CharIndex)
		// Remove from map
		s.executor.Dispatch(func(m *MapService) {
			m.RemoveCharacter(char)
		})
	}
}