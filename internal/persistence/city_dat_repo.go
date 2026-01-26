package persistence

import (
	"strconv"
	"strings"

	"github.com/ao-go-server/internal/model"
)

type CityDatRepo struct {
	path string
}

func NewCityDatRepo(path string) *CityDatRepo {
	return &CityDatRepo{path: path}
}

func (d *CityDatRepo) Load() (map[int]model.City, error) {
	data, err := ReadINI(d.path)
	if err != nil {
		return nil, err
	}

	cities := make(map[int]model.City)

	for section, props := range data {
		if !strings.HasPrefix(section, "CITY") {
			continue
		}

		id, err := strconv.Atoi(section[4:])
		if err != nil {
			continue
		}

		city := model.City{
			Map: toInt(props["MAP"]),
			X:   byte(toInt(props["X"])),
			Y:   byte(toInt(props["Y"])),
		}

		cities[id] = city
	}

	return cities, nil
}
