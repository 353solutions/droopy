package main

import (
	"fmt"
	"net"
	"sync"
)

type ConnPool struct {
	mu    sync.Mutex
	conns map[net.Conn]struct{}
}

func NewConnPool() *ConnPool {
	return &ConnPool{
		conns: make(map[net.Conn]struct{}),
	}
}

func (p *ConnPool) Add(conn net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.conns[conn] = struct{}{}
}

func (p *ConnPool) Remove(conn net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.conns, conn)
}

func (p *ConnPool) Len() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.conns)
}

func (p *ConnPool) Broadcast(msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var toRemove []net.Conn
	for conn := range p.conns {
		_, err := fmt.Fprintf(conn, "%s\n", msg)
		if err != nil {
			toRemove = append(toRemove, conn)
		}
	}

	for _, conn := range toRemove {
		delete(p.conns, conn)
		_ = conn.Close()
	}
}
