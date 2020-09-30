package server

import (
	"bufio"
	"net"
)

// a connection endpoint at server end
type conn struct {
	*bufio.Reader
	net.Conn
	server   *Server
	UserName string
	RoomName string
}

// telnet chat server
type Server struct {
	register   chan *conn
	unregister chan string
	msg        chan string
	stop       chan bool
}
