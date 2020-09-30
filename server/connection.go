package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

func newConn(s *Server, c net.Conn) *conn {
	return &conn{
		Reader: bufio.NewReader(c),
		Conn:   c,
		server: s,
	}
}

// Welcome takes care of user identity and registers
func (c *conn) Welcome() {
	var err error
	for c.UserName = ""; c.UserName == ""; {
		fmt.Fprint(c, "Enter your name: ")
		c.UserName, err = c.ReadString('\n')
		if err != nil {
			log.Printf("Reading user name from %v: %v", c.RemoteAddr(), err)
			c.Close()
			return
		}
		c.UserName = strings.TrimSpace(c.UserName)
	}

	for c.RoomName = ""; c.RoomName == ""; {
		fmt.Fprint(c, "Enter room  name: ")
		c.RoomName, err = c.ReadString('\n')
		if err != nil {
			log.Printf("Reading room name from %v: %v", c.RemoteAddr(), err)
			c.Close()
			return
		}
		c.RoomName = strings.TrimSpace(c.RoomName)
	}

	// register connection
	c.server.register <- c
}

// Readloop reads messages for this connection till its closed
func (c *conn) Readloop() {
	for {
		msg, err := c.ReadString('\n')
		if err != nil {
			break
		}

		if strings.TrimSpace(msg) != "" {
			c.server.msg <- fmt.Sprintf("%s@%s> %s", c.UserName, c.RoomName, msg)
		}
	}

	c.server.unregister <- fmt.Sprintf("%s@%s> left the room\n", c.UserName, c.RoomName)
}
