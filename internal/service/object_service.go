package service

import (
	"fmt"
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
	fmt.Println("Loading objects from data file...")
	objs, err := s.dao.Load()
	if err != nil {
		fmt.Printf("FAILED to load objects: %v\n", err)
		return err
	}
	s.objects = objs
	fmt.Printf("Successfully loaded %d objects from definitions file.\n", len(s.objects))
	return nil
}

func (s *ObjectService) GetObject(id int) *model.Object {
	return s.objects[id]
}
