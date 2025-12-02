package droopy

import (
	"bufio"
	"fmt"
	"net"
)

type Client struct {
	conn net.Conn
	scan *bufio.Scanner
}

func New(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn: conn,
		scan: bufio.NewScanner(conn),
	}, nil
}

func (c *Client) Send(cmd string) error {
	_, err := fmt.Fprintln(c.conn, cmd)
	return err
}

func (c *Client) Recv() (string, error) {
	if !c.scan.Scan() {
		if err := c.scan.Err(); err != nil {
			return "", err
		}
		return "", fmt.Errorf("connection closed")
	}

	return c.scan.Text(), nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
