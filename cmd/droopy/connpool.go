package main

import (
	"fmt"
	"maps"
	"net"
	"slices"
	"sync"
	"time"
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
	conns := slices.Collect(maps.Keys(p.conns))
	p.mu.Unlock()

	var (
		mu       sync.Mutex
		toRemove []net.Conn
		wg       sync.WaitGroup
	)

	for _, conn := range conns {
		wg.Go(func() {
			conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
			_, err := fmt.Fprintf(conn, "%s\n", msg)
			conn.SetWriteDeadline(time.Time{})
			if err != nil {
				mu.Lock()
				toRemove = append(toRemove, conn)
				mu.Unlock()
			}
		})
	}

	wg.Wait()

	p.mu.Lock()
	for _, conn := range toRemove {
		delete(p.conns, conn)
		_ = conn.Close()
	}
	p.mu.Unlock()
}
