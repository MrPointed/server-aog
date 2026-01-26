package service

import (
	"log/slog"

	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/persistence"
)

type ObjectServiceImpl struct {
	dao     persistence.ObjectRepository
	objects map[int]*model.Object
}

func NewObjectServiceImpl(dao persistence.ObjectRepository) ObjectService {
	return &ObjectServiceImpl{
		dao:     dao,
		objects: make(map[int]*model.Object),
	}
}

func (s *ObjectServiceImpl) LoadObjects() error {
	defs, err := s.dao.Load()
	if err != nil {
		slog.Error("FAILED to load objects", "error", err)
		return err
	}
	s.objects = defs
	slog.Info("Successfully loaded objects", "count", len(s.objects))
	return nil
}

func (s *ObjectServiceImpl) GetObject(id int) *model.Object {
	return s.objects[id]
}
