package persistence

import (
	"encoding/binary"
	"fmt"
	"os"
	"github.com/ao-go-server/internal/model"
)

type MapDAO struct {
	mapsPath    string
	mapsAmount  int
	waterGrhs   map[int16]bool
	lavaGrhs    map[int16]bool
}

func NewMapDAO(mapsPath string, mapsAmount int) *MapDAO {
	return &MapDAO{
		mapsPath:   mapsPath,
		mapsAmount: mapsAmount,
		waterGrhs:  make(map[int16]bool),
		lavaGrhs:   make(map[int16]bool),
	}
}

func (d *MapDAO) Load() ([]*model.Map, error) {
	maps := make([]*model.Map, d.mapsAmount)
	for i := 1; i <= d.mapsAmount; i++ {
		m, err := d.loadMap(i)
		if err != nil {
			return nil, err
		}
		maps[i-1] = m
	}
	return maps, nil
}

func (d *MapDAO) loadMap(id int) (*model.Map, error) {
	mapFileName := fmt.Sprintf("%s/Mapa%d.map", d.mapsPath, id)
	infFileName := fmt.Sprintf("%s/Mapa%d.inf", d.mapsPath, id)

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

			if (flag & BitflagLayer2) == BitflagLayer2 {
				var l2 int16
				binary.Read(mapFile, binary.LittleEndian, &l2)
			}
			if (flag & BitflagLayer3) == BitflagLayer3 {
				var l3 int16
				binary.Read(mapFile, binary.LittleEndian, &l3)
			}
			if (flag & BitflagLayer4) == BitflagLayer4 {
				var l4 int16
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
		Version: version,
		Tiles:   tiles,
	}, nil
}
