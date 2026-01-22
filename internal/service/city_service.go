package service

import (
	"fmt"
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/persistence"
)

type CityService struct {
	dao    *persistence.CityDAO
	cities map[int]model.City
}

func NewCityService(dao *persistence.CityDAO) *CityService {
	return &CityService{
		dao:    dao,
		cities: make(map[int]model.City),
	}
}

func (s *CityService) LoadCities() error {
	fmt.Println("Loading cities from data file...")

cities, err := s.dao.Load()
	if err != nil {
		return err
	}
	s.cities = cities
	fmt.Printf("Successfully loaded %d cities.\n", len(s.cities))
	return nil
}

func (s *CityService) GetCity(id int) (model.City, bool) {
	city, ok := s.cities[id]
	return city, ok
}
