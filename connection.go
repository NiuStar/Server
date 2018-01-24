package server

import (
	"fmt"
	"net"
)

type Connection struct {
	net.Conn
	listener *Listener

	closed bool
}

func (this *Connection) Close() error {
	fmt.Println("close1")
	if !this.closed {
		this.closed = true
		this.listener.waitGroup.Done()
	}
	fmt.Println("close")
	return this.Conn.Close()
}
