package service

import (
	"log/slog"

	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/persistence"
)

type ObjectService struct {
	dao     *persistence.ObjectDAO
	objects map[int]*model.Object
}

func NewObjectService(dao *persistence.ObjectDAO) *ObjectService {
	return &ObjectService{
		dao:     dao,
		objects: make(map[int]*model.Object),
	}
}

func (s *ObjectService) LoadObjects() error {
	defs, err := s.dao.Load()
	if err != nil {
		slog.Error("FAILED to load objects", "error", err)
		return err
	}
	s.objects = defs
	slog.Info("Successfully loaded objects", "count", len(s.objects))
	return nil
}

func (s *ObjectService) GetObject(id int) *model.Object {
	return s.objects[id]
}
