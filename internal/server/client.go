package server

import (
	"net"
	"sync"
	"time"
)

type Client struct {
	conn     net.Conn
	name     string
	joinTime time.Time
	mu       sync.Mutex
}

func NewClient(conn net.Conn) *Client {
	return &Client{
		conn:     conn,
		joinTime: time.Now(),
	}
}

func (c *Client) GetName() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.name
}

func (c *Client) SetName(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.name = name
}

func (c *Client) SendMessage(message string) error {
	_, err := c.conn.Write([]byte(message))
	return err
}

func (c *Client) GetAddress() string {
	return c.conn.RemoteAddr().String()
}
