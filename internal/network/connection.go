package network

import (
	"net"
	"github.com/ao-go-server/internal/model"
)

type Connection struct {
	Conn net.Conn
	User *model.Character
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		Conn: conn,
	}
}

func (c *Connection) Disconnect() {
	c.Conn.Close()
}

// Send will be implemented once we have the protocol package ready to avoid circular dependencies if needed, 
// or we can use the interface defined in protocol.
// For now, let's just leave it placeholder or use a raw byte sender.
func (c *Connection) SendBytes(data []byte) error {
	_, err := c.Conn.Write(data)
	return err
}
