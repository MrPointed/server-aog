package service

import (
	"sync"
)

type CharacterIndexManager struct {
	nextIndex int16
	freePool  []int16
	mu        sync.Mutex
}

func NewCharacterIndexManager() *CharacterIndexManager {
	return &CharacterIndexManager{
		nextIndex: 0,
		freePool:  make([]int16, 0),
	}
}

func (m *CharacterIndexManager) AssignIndex() int16 {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Use from pool if available
	if len(m.freePool) > 0 {
		index := m.freePool[len(m.freePool)-1]
		m.freePool = m.freePool[:len(m.freePool)-1]
		return index
	}

	// Otherwise increment
	m.nextIndex++
	return m.nextIndex
}

func (m *CharacterIndexManager) FreeIndex(index int16) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.freePool = append(m.freePool, index)
}
