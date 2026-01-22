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

type options struct {
	addr string
}

// ClientOption is a function that configures a Client.
type ClientOption func(*options)

// WithAddr sets the server address for the client.
func WithAddr(addr string) ClientOption {
	return func(o *options) {
		o.addr = addr
	}
}

// NewClient return new client connected to simulator.
func NewClient(opts ...ClientOption) (*Client, error) {
	o := options{
		addr: "localhost:10000",
	}

	for _, opt := range opts {
		opt(&o)
	}

	conn, err := net.Dial("tcp", o.addr)
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

// Recv receives an event from the simulator, blocking until there's one.
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
