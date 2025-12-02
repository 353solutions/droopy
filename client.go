package droopy

import (
	"bufio"
	"fmt"
	"net"
)

// Client is a client to the simulator.
type Client struct {
	conn net.Conn
	scan *bufio.Scanner
}

// NewClient return new client connected to simulator at `addr`.
func NewClient(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn: conn,
		scan: bufio.NewScanner(conn),
	}, nil
}

// Send sends a command to the client.
func (c *Client) Send(cmd string) error {
	_, err := fmt.Fprintln(c.conn, cmd)
	return err
}

// Recv receives an event from the simulator, blocking until there's on.
func (c *Client) Recv() (string, error) {
	if !c.scan.Scan() {
		if err := c.scan.Err(); err != nil {
			return "", err
		}
		return "", fmt.Errorf("connection closed")
	}

	return c.scan.Text(), nil
}

// Close closes the client.
func (c *Client) Close() error {
	return c.conn.Close()
}
