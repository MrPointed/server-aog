package network

import (
	"encoding/binary"
	"errors"
	"math"
)

type DataBuffer struct {
	buf []byte
	pos int
}

func NewDataBuffer(data []byte) *DataBuffer {
	return &DataBuffer{
		buf: data,
		pos: 0,
	}
}

func (db *DataBuffer) Bytes() []byte {
	return db.buf
}

func (db *DataBuffer) ReadableBytes() int {
	return len(db.buf) - db.pos
}

func (db *DataBuffer) Pos() int {
	return db.pos
}

func (db *DataBuffer) SetPos(pos int) {
	db.pos = pos
}

func (db *DataBuffer) Get() (byte, error) {
	if db.ReadableBytes() < 1 {
		return 0, errors.New("not enough data")
	}
	b := db.buf[db.pos]
	db.pos++
	return b, nil
}

func (db *DataBuffer) GetBoolean() (bool, error) {
	b, err := db.Get()
	if err != nil {
		return false, err
	}
	return b == 1, nil
}

func (db *DataBuffer) GetShort() (int16, error) {
	if db.ReadableBytes() < 2 {
		return 0, errors.New("not enough data")
	}
	v := int16(binary.LittleEndian.Uint16(db.buf[db.pos:]))
	db.pos += 2
	return v, nil
}

func (db *DataBuffer) GetInt() (int32, error) {
	if db.ReadableBytes() < 4 {
		return 0, errors.New("not enough data")
	}
	v := int32(binary.LittleEndian.Uint32(db.buf[db.pos:]))
	db.pos += 4
	return v, nil
}

func (db *DataBuffer) GetLong() (int64, error) {
	if db.ReadableBytes() < 8 {
		return 0, errors.New("not enough data")
	}
	v := int64(binary.LittleEndian.Uint64(db.buf[db.pos:]))
	db.pos += 8
	return v, nil
}

func (db *DataBuffer) GetFloat() (float32, error) {
	if db.ReadableBytes() < 4 {
		return 0, errors.New("not enough data")
	}
	bits := binary.LittleEndian.Uint32(db.buf[db.pos:])
	v := math.Float32frombits(bits)
	db.pos += 4
	return v, nil
}

func (db *DataBuffer) GetDouble() (float64, error) {
	if db.ReadableBytes() < 8 {
		return 0, errors.New("not enough data")
	}
	bits := binary.LittleEndian.Uint64(db.buf[db.pos:])
	v := math.Float64frombits(bits)
	db.pos += 8
	return v, nil
}

func (db *DataBuffer) GetUTF8String() (string, error) {
	length, err := db.GetShort()
	if err != nil {
		return "", err
	}
	if db.ReadableBytes() < int(length) {
		return "", errors.New("not enough data")
	}
	s := string(db.buf[db.pos : db.pos+int(length)])
	db.pos += int(length)
	return s, nil
}

func (db *DataBuffer) GetUTF8StringFixed(length int) (string, error) {
	if db.ReadableBytes() < length {
		return "", errors.New("not enough data")
	}
	data := db.buf[db.pos : db.pos+length]
	db.pos += length

	// Find the real end of string (remove trailing null bytes)
	actualLength := length
	for i := length - 1; i >= 0; i-- {
		if data[i] != 0 {
			actualLength = i + 1
			break
		}
	}
	if actualLength == 0 {
		return "", nil
	}
	return string(data[:actualLength]), nil
}

func (db *DataBuffer) Put(b byte) {
	db.buf = append(db.buf, b)
}

func (db *DataBuffer) PutBytes(data []byte) {
	db.buf = append(db.buf, data...)
}

func (db *DataBuffer) PutBoolean(v bool) {
	if v {
		db.Put(1)
	} else {
		db.Put(0)
	}
}

func (db *DataBuffer) PutShort(v int16) {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(v))
	db.buf = append(db.buf, b...)
}

func (db *DataBuffer) PutInt(v int32) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(v))
	db.buf = append(db.buf, b...)
}

func (db *DataBuffer) PutLong(v int64) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(v))
	db.buf = append(db.buf, b...)
}

func (db *DataBuffer) PutFloat(v float32) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, math.Float32bits(v))
	db.buf = append(db.buf, b...)
}

func (db *DataBuffer) PutDouble(v float64) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, math.Float64bits(v))
	db.buf = append(db.buf, b...)
}

func (db *DataBuffer) PutUTF8String(s string) {
	b := []byte(s)
	db.PutShort(int16(len(b)))
	db.buf = append(db.buf, b...)
}

func (db *DataBuffer) PutCp1252String(s string) {
	// For now, AO Go server seems to just use raw bytes for string.
	// We'll treat it as standard string bytes which is usually compatible for ASCII.
	b := []byte(s)
	db.PutShort(int16(len(b)))
	db.buf = append(db.buf, b...)
}