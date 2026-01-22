package service

import (
	"sync"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/model"
)

type UserService struct {
	loggedUsers        map[protocol.Connection]bool
	charToConn         map[*model.Character]protocol.Connection
	loggedCharsByIndex map[int16]*model.Character
	mu                 sync.RWMutex
}

func NewUserService() *UserService {
	return &UserService{
		loggedUsers:        make(map[protocol.Connection]bool),
		charToConn:         make(map[*model.Character]protocol.Connection),
		loggedCharsByIndex: make(map[int16]*model.Character),
	}
}

func (s *UserService) IsLoggedIn(conn protocol.Connection) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.loggedUsers[conn]
}

func (s *UserService) LogIn(conn protocol.Connection) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.loggedUsers[conn] = true
	char := conn.GetUser()
	if char != nil {
		s.charToConn[char] = conn
		s.loggedCharsByIndex[char.CharIndex] = char
	}
}

func (s *UserService) LogOut(conn protocol.Connection) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.loggedUsers, conn)
	char := conn.GetUser()
	if char != nil {
		delete(s.charToConn, char)
		delete(s.loggedCharsByIndex, char.CharIndex)
	}
}

func (s *UserService) GetCharacterByIndex(index int16) *model.Character {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.loggedCharsByIndex[index]
}

func (s *UserService) GetConnection(char *model.Character) protocol.Connection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.charToConn[char]
}

func (s *UserService) GetLoggedCharacters() []*model.Character {
	s.mu.RLock()
	defer s.mu.RUnlock()
	chars := make([]*model.Character, 0, len(s.charToConn))
	for char := range s.charToConn {
		chars = append(chars, char)
	}
	return chars
}
