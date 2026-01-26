package persistence

import (
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ao-go-server/internal/model"
	"gopkg.in/yaml.v3"
)

type MapDatRepo struct {
	mapsPath   string
	mapsAmount int
	waterGrhs  map[int16]bool
	lavaGrhs   map[int16]bool
}

func NewMapDatRepo(mapsPath string, mapsAmount int) *MapDatRepo {
	return &MapDatRepo{
		mapsPath:   mapsPath,
		mapsAmount: mapsAmount,
		waterGrhs:  make(map[int16]bool),
		lavaGrhs:   make(map[int16]bool),
	}
}

func (d *MapDatRepo) GetMapsAmount() int {
	return d.mapsAmount
}

type yamlMaps struct {
	Maps struct {
		Tiles struct {
			Water []string `yaml:"water"`
			Lava  []string `yaml:"lava"`
		} `yaml:"tiles"`
	} `yaml:"maps"`
}

func (d *MapDatRepo) LoadProperties(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var ym yamlMaps
	if err := yaml.Unmarshal(data, &ym); err != nil {
		return err
	}

	for _, r := range ym.Maps.Tiles.Water {
		d.parseRanges(r, d.waterGrhs)
	}
	for _, r := range ym.Maps.Tiles.Lava {
		d.parseRanges(r, d.lavaGrhs)
	}

	return nil
}

func (d *MapDatRepo) parseRanges(r string, target map[int16]bool) {
	r = strings.TrimSpace(r)
	bounds := strings.Split(r, "-")
	if len(bounds) == 2 {
		start, err1 := strconv.Atoi(strings.TrimSpace(bounds[0]))
		end, err2 := strconv.Atoi(strings.TrimSpace(bounds[1]))
		if err1 == nil && err2 == nil {
			for i := start; i <= end; i++ {
				target[int16(i)] = true
			}
		}
	} else {
		// Single value
		val, err := strconv.Atoi(r)
		if err == nil {
			target[int16(val)] = true
		}
	}
}

func (d *MapDatRepo) Load() ([]*model.Map, error) {
	maps := make([]*model.Map, d.mapsAmount)
	for i := 1; i <= d.mapsAmount; i++ {
		m, err := d.LoadMap(i)
		if err != nil {
			return nil, err
		}
		maps[i-1] = m
	}
	return maps, nil
}

func (d *MapDatRepo) LoadMap(id int) (*model.Map, error) {
	mapFileName := fmt.Sprintf("%s/Mapa%d.map", d.mapsPath, id)
	infFileName := fmt.Sprintf("%s/Mapa%d.inf", d.mapsPath, id)
	datFileName := fmt.Sprintf("%s/Mapa%d.dat", d.mapsPath, id)

	// Check if .dat file exists with alternative casing (some files are mapa7.dat)
	if _, err := os.Stat(datFileName); os.IsNotExist(err) {
		datFileName = fmt.Sprintf("%s/mapa%d.dat", d.mapsPath, id)
	}

	mapFile, err := os.Open(mapFileName)
	if err != nil {
		return nil, err
	}
	defer mapFile.Close()

	infFile, err := os.Open(infFileName)
	if err != nil {
		return nil, err
	}
	defer infFile.Close()

	// Load properties from .dat
	pkMap := true // Default to PK allowed
	mapName := ""
	if datProps, err := ReadINI(datFileName); err == nil {
		sectionKey := fmt.Sprintf("MAPA%d", id)
		if header, ok := datProps[sectionKey]; ok {
			if pkVal, ok := header["PK"]; ok {
				pkMap = pkVal == "0"
			}
			if nameVal, ok := header["NOMBRE"]; ok {
				mapName = nameVal
			}
		} else if header, ok := datProps["MAPA"]; ok {
			if pkVal, ok := header["PK"]; ok {
				pkMap = pkVal == "0"
			}
			if nameVal, ok := header["NOMBRE"]; ok {
				mapName = nameVal
			}
		} else if header, ok := datProps["MAIN"]; ok {
			if pkVal, ok := header["PK"]; ok {
				pkMap = pkVal == "0"
			}
		}
	}

	// Header Map
	var version int16
	binary.Read(mapFile, binary.LittleEndian, &version)

	description := make([]byte, 255)
	mapFile.Read(description)

	var crc int32
	binary.Read(mapFile, binary.LittleEndian, &crc)
	var magic int32
	binary.Read(mapFile, binary.LittleEndian, &magic)

	var unusedLong int64
	binary.Read(mapFile, binary.LittleEndian, &unusedLong)

	// Header Inf
	binary.Read(infFile, binary.LittleEndian, &unusedLong)
	var unusedShort int16
	binary.Read(infFile, binary.LittleEndian, &unusedShort)

	tiles := make([]model.Tile, model.MapWidth*model.MapHeight)

	const (
		BitflagBlocked = 1
		BitflagLayer2  = 2
		BitflagLayer3  = 4
		BitflagLayer4  = 8
		BitflagTrigger = 16

		BitflagTileExit = 1
		BitflagNpc      = 2
		BitflagObject   = 4
	)

	for y := 0; y < model.MapHeight; y++ {
		for x := 0; x < model.MapWidth; x++ {
			var flag byte
			binary.Read(mapFile, binary.LittleEndian, &flag)

			blocked := (flag & BitflagBlocked) == BitflagBlocked

			var floor int16
			binary.Read(mapFile, binary.LittleEndian, &floor)

			isWater := d.waterGrhs[floor]
			isLava := d.lavaGrhs[floor]

			var l2, l3, l4 int16

			if (flag & BitflagLayer2) == BitflagLayer2 {
				binary.Read(mapFile, binary.LittleEndian, &l2)
			}
			if (flag & BitflagLayer3) == BitflagLayer3 {
				binary.Read(mapFile, binary.LittleEndian, &l3)
			}
			if (flag & BitflagLayer4) == BitflagLayer4 {
				binary.Read(mapFile, binary.LittleEndian, &l4)
			}

			var trigger model.Trigger = model.TriggerNone
			if (flag & BitflagTrigger) == BitflagTrigger {
				var trigIdx int16
				binary.Read(mapFile, binary.LittleEndian, &trigIdx)
				trigger = model.Trigger(trigIdx)
			}

			// Info file
			var infFlag byte
			binary.Read(infFile, binary.LittleEndian, &infFlag)

			var tileExit *model.Position
			if (infFlag & BitflagTileExit) == BitflagTileExit {
				var toMap int16
				var toX, toY int16
				binary.Read(infFile, binary.LittleEndian, &toMap)
				binary.Read(infFile, binary.LittleEndian, &toX)
				binary.Read(infFile, binary.LittleEndian, &toY)
				tileExit = &model.Position{X: byte(toX - 1), Y: byte(toY - 1), Map: int(toMap)}
			}

			var npcIdx int16
			if (infFlag & BitflagNpc) == BitflagNpc {
				binary.Read(infFile, binary.LittleEndian, &npcIdx)
			}

			var objIdx, objAmount int16
			if (infFlag & BitflagObject) == BitflagObject {
				binary.Read(infFile, binary.LittleEndian, &objIdx)
				binary.Read(infFile, binary.LittleEndian, &objAmount)
			}

			tiles[y*model.MapWidth+x] = model.Tile{
				Blocked:      blocked,
				IsWater:      isWater,
				IsLava:       isLava,
				Layer2:       l2,
				Layer3:       l3,
				Layer4:       l4,
				Trigger:      trigger,
				TileExit:     tileExit,
				ObjectID:     int(objIdx),
				ObjectAmount: int(objAmount),
				NPCID:        int(npcIdx),
			}
		}
	}

	return &model.Map{
		Id:      id,
		Name:    mapName,
		Version: version,
		Pk:      pkMap,
		Tiles:   tiles,
	}, nil
}
