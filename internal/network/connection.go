package network

import (
	"net"
	"github.com/ao-go-server/internal/model"
)

type Connection struct {
	Conn       net.Conn
	Attributes map[int]byte
	User       *model.Character
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		Conn:       conn,
		Attributes: make(map[int]byte),
	}
}

func (c *Connection) SendBytes(data []byte) error {
	_, err := c.Conn.Write(data)
	return err
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
