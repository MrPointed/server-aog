package service

import "sync/atomic"

type CharacterIndexManager struct {
	nextIndex int32
}

func NewCharacterIndexManager() *CharacterIndexManager {
	return &CharacterIndexManager{
		nextIndex: 0,
	}
}

func (m *CharacterIndexManager) AssignIndex() int16 {
	return int16(atomic.AddInt32(&m.nextIndex, 1))
}
