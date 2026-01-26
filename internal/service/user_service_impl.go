package service

import (
	"sync"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/model"
)

type UserServiceImpl struct {
	loggedUsers        map[protocol.Connection]bool
	charToConn         map[*model.Character]protocol.Connection
	loggedCharsByIndex map[int16]*model.Character
	mu                 sync.RWMutex
	bodyService        BodyService
}

func NewUserServiceImpl(bodyService BodyService) UserService {
	return &UserServiceImpl{
		loggedUsers:        make(map[protocol.Connection]bool),
		charToConn:         make(map[*model.Character]protocol.Connection),
		loggedCharsByIndex: make(map[int16]*model.Character),
		bodyService:        bodyService,
	}
}

func (s *UserServiceImpl) BodyService() BodyService {
	return s.bodyService
}

func (s *UserServiceImpl) IsLoggedIn(conn protocol.Connection) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.loggedUsers[conn]
}

func (s *UserServiceImpl) IsUserLoggedIn(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for char := range s.charToConn {
		if char.Name == name {
			return true
		}
	}
	return false
}

func (s *UserServiceImpl) GetCharacterByName(name string) *model.Character {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for char := range s.charToConn {
		if char.Name == name {
			return char
		}
	}
	return nil
}

func (s *UserServiceImpl) LogIn(conn protocol.Connection) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.loggedUsers[conn] = true
	char := conn.GetUser()
	if char != nil {
		s.charToConn[char] = conn
		s.loggedCharsByIndex[char.CharIndex] = char
	}
}

func (s *UserServiceImpl) LogOut(conn protocol.Connection) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.loggedUsers, conn)
	char := conn.GetUser()
	if char != nil {
		delete(s.charToConn, char)
		delete(s.loggedCharsByIndex, char.CharIndex)
	}
}

func (s *UserServiceImpl) GetCharacterByIndex(index int16) *model.Character {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.loggedCharsByIndex[index]
}

func (s *UserServiceImpl) GetConnection(char *model.Character) protocol.Connection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.charToConn[char]
}

func (s *UserServiceImpl) GetLoggedCharacters() []*model.Character {
	s.mu.RLock()
	defer s.mu.RUnlock()
	chars := make([]*model.Character, 0, len(s.charToConn))
	for char := range s.charToConn {
		chars = append(chars, char)
	}
	return chars
}

func (s *UserServiceImpl) GetLoggedConnections() []protocol.Connection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	conns := make([]protocol.Connection, 0, len(s.loggedUsers))
	for conn := range s.loggedUsers {
		conns = append(conns, conn)
	}
	return conns
}

func (s *UserServiceImpl) KickByName(name string) bool {
	char := s.GetCharacterByName(name)
	if char == nil {
		return false
	}
	conn := s.GetConnection(char)
	if conn == nil {
		return false
	}
	conn.Disconnect()
	return true
}

func (s *UserServiceImpl) KickByIP(ip string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	kicked := 0
	for conn := range s.loggedUsers {
		if conn.GetRemoteAddr() == ip {
			conn.Disconnect()
			kicked++
		}
	}
	return kicked
}
