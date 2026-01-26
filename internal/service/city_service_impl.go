package service

import (

	"log/slog"



	"github.com/ao-go-server/internal/model"

	"github.com/ao-go-server/internal/persistence"

)



type CityServiceImpl struct {

	dao    persistence.CityRepository

	cities map[int]model.City

}



func NewCityServiceImpl(dao persistence.CityRepository) CityService {

	return &CityServiceImpl{

		dao:    dao,

		cities: make(map[int]model.City),

	}

}



func (s *CityServiceImpl) LoadCities() error {

	slog.Info("Loading cities from data file...")

	cities, err := s.dao.Load()

	if err != nil {

		return err

	}

	s.cities = cities

	slog.Info("Successfully loaded cities", "count", len(s.cities))

	return nil

}



func (s *CityServiceImpl) GetCity(id int) (model.City, bool) {
	city, ok := s.cities[id]
	return city, ok
}
