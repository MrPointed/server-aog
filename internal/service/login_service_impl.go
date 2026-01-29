package service

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/ao-go-server/internal/config"
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/persistence"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type LoginServiceImpl struct {
	userRepo       persistence.UserRepository
	config         *config.Config
	projectConfig  *config.ProjectConfig
	userService    UserService
	mapService     MapService
	bodyService    BodyService
	indexManager   *CharacterIndexManager
	messageService MessageService
	objectService  ObjectService
	cityService    CityService
	spellService   SpellService
}

func NewLoginServiceImpl(userRepo persistence.UserRepository,
	cfg *config.Config, projectCfg *config.ProjectConfig, userService UserService, mapService MapService,
	bodyService BodyService, indexManager *CharacterIndexManager,
	messageService MessageService, objectService ObjectService, cityService CityService,
	spellService SpellService) LoginService {
	return &LoginServiceImpl{
		userRepo:       userRepo,
		config:         cfg,
		projectConfig:  projectCfg,
		userService:    userService,
		mapService:     mapService,
		bodyService:    bodyService,
		indexManager:   indexManager,
		messageService: messageService,
		objectService:  objectService,
		cityService:    cityService,
		spellService:   spellService,
	}
}

// ConnectNewCharacter handles the creation and login of a new character.
func (s *LoginServiceImpl) ConnectNewCharacter(conn protocol.Connection, nick, password, mail string,
	raceId, genderId, archetypeId, headId, cityId byte, clientHash, version string) error {
	slog.Info("Connecting new character", "nick", nick)

	if err := s.validateLoginRequest(nick, version, clientHash); err != nil {
		return err
	}

	if !s.config.CharacterCreationEnabled {
		return fmt.Errorf("la creación de personajes está deshabilitada")
	}

	// Check if name taken
	if s.userRepo.Exists(nick) {
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

	acc, char, err := s.userRepo.CreateAccountAndCharacter(nick, password, mail,
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
func (s *LoginServiceImpl) ConnectExistingCharacter(conn protocol.Connection, nick, password, version, clientHash string) error {
	slog.Info("Connecting existing character", "nick", nick)

	// 1. Basic Validation (Version & Name)
	if err := s.validateLoginRequest(nick, version, clientHash); err != nil {
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
	char, err := s.userRepo.Load(nick)
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

// validateLoginRequest checks version, name validity and MD5 hash.
func (s *LoginServiceImpl) validateLoginRequest(nick, version, clientHash string) error {
	if s.config.Version != version {
		return fmt.Errorf("versión obsoleta, por favor actualice el juego")
	}
	if nick == "" {
		return fmt.Errorf("nombre inválido")
	}

	if s.config.MD5Enabled {
		valid := false
		for _, h := range s.config.AcceptedMD5s {
			if h == clientHash {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("cliente no autorizado (hash inválido)")
		}
	}

	return nil
}

func (s *LoginServiceImpl) checkServerCapacity() error {
	currentUsers := len(s.userService.GetLoggedCharacters())
	if currentUsers >= s.config.MaxConcurrentUsers {
		return fmt.Errorf("el servidor está lleno, intente más tarde")
	}
	return nil
}

// authenticate verifies account credentials.
func (s *LoginServiceImpl) authenticate(nick, password string) (*model.Account, error) {
	acc, err := s.userRepo.Get(nick)
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
func (s *LoginServiceImpl) checkBan(acc *model.Account) error {
	if acc.Banned {
		return fmt.Errorf("acceso denegado: cuenta suspendida")
	}
	return nil
}

// checkAlreadyLoggedIn ensures the character is not already in the game.
func (s *LoginServiceImpl) checkAlreadyLoggedIn(nick string) error {
	if s.userService.IsUserLoggedIn(nick) {
		return fmt.Errorf("el personaje ya se encuentra conectado")
	}
	return nil
}

// ensureValidCharacterState sets the correct body and head depending on dead/alive state.
func (s *LoginServiceImpl) ensureValidCharacterState(char *model.Character, defaultBody int) {
	if char.Dead {
		char.Body = 8
		char.Head = 500
	} else if char.Body == 0 {
		char.Body = defaultBody
	}
}

func (s *LoginServiceImpl) finalizeLogin(conn protocol.Connection, acc *model.Account, char *model.Character) {
	s.determinePrivileges(char)
	conn.Send(&outgoing.LoggedPacket{})

	// Assign Index
	char.CharIndex = s.indexManager.AssignIndex()

	// Initial skills if new character
	if char.Level == 1 && char.Exp == 0 {
		char.SkillPoints = s.projectConfig.Project.LoginService.InitialAvailableSkills
	}

	// Bind to connection and service
	conn.SetUser(char)
	s.userService.LogIn(conn)

	// Dispatch to World (Thread-safe map modification)
	// TODO: Add collision check logic here similar to 'LegalPos' in VB6
	s.mapService.PutCharacterAtPos(char, char.Position)

	// Send Handshake / Game State
	s.sendInitialGameState(conn, char)

	// Notify others
	s.messageService.SendToAreaButUser(&outgoing.CharacterCreatePacket{Character: char}, char.Position, char)
	s.messageService.AreaService().SendAreaState(char)

	slog.Info("User logged in", "name", char.Name, "pos", char.Position, "privs", char.Privileges)
}

func (s *LoginServiceImpl) determinePrivileges(char *model.Character) {
	name := strings.ToUpper(char.Name)
	char.Privileges = model.PrivilegeUser

	// Check in descending order of power
	for _, admin := range s.config.Gods {
		if name == admin {
			char.Privileges = model.PrivilegeGod
			return
		}
	}
	for _, admin := range s.config.SemiGods {
		if name == admin {
			char.Privileges = model.PrivilegeSemiGod
			return
		}
	}
	for _, admin := range s.config.Counselors {
		if name == admin {
			char.Privileges = model.PrivilegeCounselor
			return
		}
	}
	for _, admin := range s.config.RoleMasters {
		if name == admin {
			char.Privileges = model.PrivilegeRoleMaster
			return
		}
	}
}

func (s *LoginServiceImpl) sendInitialGameState(conn protocol.Connection, char *model.Character) {
	// Send UserIndexInServer (Packet 27) - Contains Privileges & Color
	conn.Send(&outgoing.UserIndexInServerPacket{
		UserIndex: char.CharIndex,
	})

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

	if char.Sailing {
		conn.Send(&outgoing.NavigateTogglePacket{})
	}

	if char.Meditating {
		conn.Send(&outgoing.MeditateTogglePacket{})
		fx := &outgoing.CreateFxPacket{
			CharIndex: char.CharIndex,
			FxID:      4,
			Loops:     -1,
		}
		conn.Send(fx)
		s.messageService.SendToAreaButUser(fx, char.Position, char)
	}

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

func (s *LoginServiceImpl) getChatColor(privs model.PrivilegeLevel) int32 {
	rgb := func(r, g, b int32) int32 {
		return r + (g * 256) + (b * 65536)
	}

	if privs.IsGod() {
		return rgb(250, 250, 150)
	} else if privs == model.PrivilegeCounselor {
		return rgb(0, 255, 0)
	} else if privs == model.PrivilegeSemiGod {
		return rgb(0, 255, 0)
	}
	// Default white
	return rgb(255, 255, 255)
}

func (s *LoginServiceImpl) sendInventory(conn protocol.Connection, char *model.Character) {
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

func (s *LoginServiceImpl) sendSpells(conn protocol.Connection, char *model.Character) {
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

func (s *LoginServiceImpl) OnUserDisconnect(conn protocol.Connection) {
	char := conn.GetUser()
	if char != nil {
		slog.Info("User disconnected, saving...", "name", char.Name)
		s.SavePlayer(char.Name)

		// Broadcast removal
		s.messageService.SendToAreaButUser(&outgoing.CharacterRemovePacket{CharIndex: char.CharIndex}, char.Position, char)

		s.userService.LogOut(conn)
		s.indexManager.FreeIndex(char.CharIndex)
		// Remove from map
		s.mapService.RemoveCharacter(char)
	}
}

func (s *LoginServiceImpl) LockAccount(nick string) error {
	acc, err := s.userRepo.Get(nick)
	if err != nil {
		return err
	}
	acc.Banned = true
	return s.userRepo.SaveAccount(acc)
}

func (s *LoginServiceImpl) UnlockAccount(nick string) error {
	acc, err := s.userRepo.Get(nick)
	if err != nil {
		return err
	}
	acc.Banned = false
	return s.userRepo.SaveAccount(acc)
}

func (s *LoginServiceImpl) ResetPassword(nick string, newPassword string) error {
	acc, err := s.userRepo.Get(nick)
	if err != nil {
		return err
	}
	acc.Password = newPassword
	return s.userRepo.SaveAccount(acc)
}

func (s *LoginServiceImpl) TeleportPlayer(nick string, mapID, x, y int) error {
	char := s.userService.GetCharacterByName(nick)
	if char == nil {
		// If not online, we could theoretically modify the saved file,
		// but usually teleport is for online players.
		return fmt.Errorf("player not online")
	}

	if !s.mapService.IsInPlayableArea(x, y) {
		return fmt.Errorf("invalid position")
	}

	newPos := model.Position{Map: mapID, X: byte(x), Y: byte(y)}
	s.mapService.PutCharacterAtPos(char, newPos)
	// Notify the client and surrounding areas
	s.messageService.SendToAreaButUser(&outgoing.CharacterRemovePacket{CharIndex: char.CharIndex}, char.Position, char)
	conn := s.userService.GetConnection(char)
	if conn != nil {
		conn.Send(&outgoing.AreaChangedPacket{Position: newPos})
	}
	s.messageService.SendToAreaButUser(&outgoing.CharacterCreatePacket{Character: char}, newPos, char)

	return nil
}

func (s *LoginServiceImpl) SavePlayer(nick string) error {
	char := s.userService.GetCharacterByName(nick)
	if char != nil {
		return s.userRepo.SaveCharacter(char)
	}
	return fmt.Errorf("player not found or not online")
}

func (s *LoginServiceImpl) SaveAllPlayers() {
	chars := s.userService.GetLoggedCharacters()
	for _, char := range chars {
		s.userRepo.SaveCharacter(char)
	}
}

func (s *LoginServiceImpl) WorldSave() {
	slog.Info("Starting WorldSave...")
	s.messageService.BroadcastMessage("Servidor> Iniciando WorldSave", outgoing.SERVER)

	// Save all logged players
	s.SaveAllPlayers()

	// Save map cache (includes objects on ground and map state)
	s.mapService.SaveCache()

	s.messageService.BroadcastMessage("Servidor> WorldSave ha concluido.", outgoing.SERVER)
	slog.Info("WorldSave completed.")

	// Log WorldSave
	logFile := "logs/BackUps.log"
	os.MkdirAll("logs", 0755)
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		timestamp := time.Now().Format("02/01/2006 15:04:05")
		f.WriteString(timestamp + "\n")
	}
}
