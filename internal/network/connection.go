package network

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/ao-go-server/internal/model"
)

type Connection struct {
	Conn       net.Conn
	Attributes map[int]byte
	User       *model.Character
	
	// Stats
	BytesIn    uint64
	BytesOut   uint64
	PacketsIn  uint64
	PacketsOut uint64
	StartTime  time.Time
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		Conn:       conn,
		Attributes: make(map[int]byte),
		StartTime:  time.Now(),
	}
}

func (c *Connection) SendBytes(data []byte) error {
	n, err := c.Conn.Write(data)
	if n > 0 {
		c.UpdateStats(n, false)
	}
	return err
}

func (c *Connection) UpdateStats(n int, isInput bool) {
	if isInput {
		atomic.AddUint64(&c.BytesIn, uint64(n))
		atomic.AddUint64(&c.PacketsIn, 1)
	} else {
		atomic.AddUint64(&c.BytesOut, uint64(n))
		atomic.AddUint64(&c.PacketsOut, 1)
	}
}

func (c *Connection) GetStats() (in uint64, out uint64, pIn uint64, pOut uint64, start time.Time) {
	return atomic.LoadUint64(&c.BytesIn), atomic.LoadUint64(&c.BytesOut), atomic.LoadUint64(&c.PacketsIn), atomic.LoadUint64(&c.PacketsOut), c.StartTime
}

func (c *Connection) Disconnect() {
	c.Conn.Close()
}

func (c *Connection) GetRemoteAddr() string {
	return c.Conn.RemoteAddr().String()
}

func (c *Connection) SetAttribute(attr int, value byte) {
	if c.Attributes == nil {
		c.Attributes = make(map[int]byte)
	}
	c.Attributes[attr] = value
}

func (c *Connection) GetAttribute(attr int) byte {
	if c.Attributes == nil {
		return 0
	}
	return c.Attributes[attr]
}

func (c *Connection) GetUser() *model.Character {
	return c.User
}

func (c *Connection) SetUser(user *model.Character) {
	c.User = user
}
