package server

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"time"
)

var ip = flag.String("ip", "127.0.0.1", "IP address to listen on")
var port = flag.String("port", "4000", "Port for server to listen on")

// ListenAndServe ... you guess :-)
func ListenAndServe(addr string) error {
	// Resolve the passed port into an address
	addrs, err := net.ResolveTCPAddr("tcp", *ip+":"+*port)
	if err != nil {
		return err
	}

	ln, err := net.ListenTCP("tcp", addrs)
	if err != nil {
		return err
	}
	log.Println("Listening for connections on: ", addrs)
	defer ln.Close()
	s := &Server{
		register:   make(chan *conn),
		unregister: make(chan string),
		msg:        make(chan string),
		stop:       make(chan bool),
	}

	// handle connections as they come
	go s.handleConns()

	for {
		rwc, err := ln.Accept()
		if err != nil {
			close(s.stop)
			return err
		}
		log.Println("New connection from: ", rwc.RemoteAddr())
		go newConn(s, rwc).Welcome()
	}
}

func (s *Server) handleConns() {
	type userInfo struct {
		userName string
		userConn net.Conn
	}

	// set of user names in use
	usrNames := make(map[string]interface{})
	// room to users map
	roomsToUserInfos := make(map[string][]userInfo)
	roomsToFiles := make(map[string]*os.File)

	broadcast := func(str string) {
		userName, roomName, _ := parseMsg(str)
		if (userName != "") && (roomName != "") {
			log.Printf("%s", str)
			if uInfos, found := roomsToUserInfos[roomName]; found {
				for _, uInfo := range uInfos {
					_, c := uInfo.userName, uInfo.userConn
					c.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
					_, _ = c.Write([]byte(str))
				}
			} else {
				log.Println("Room or User info not found")
			}

			f := roomsToFiles[roomName]
			f.Write([]byte(str))
			f.Sync()
		}
	}

	regConn := func(c *conn) bool {
		// if user name exists, ask to choose unique user name
		if _, exists := usrNames[c.UserName]; exists {
			fmt.Fprintf(c, "Name %q already in use.\n", c.UserName)
			go c.Welcome()
			// c.Close()
			return false
		}

		// add user name to set of names
		usrNames[c.UserName] = nil

		// if room is non-existent, create one
		if _, exists := roomsToUserInfos[c.RoomName]; !exists {
			ui := userInfo{c.UserName, c.Conn}
			uInfos := make([]userInfo, 0)
			roomsToUserInfos[c.RoomName] = append(uInfos, ui)

			fname := fmt.Sprintf("chatroom_%s.log", c.RoomName)
			f, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE, 0755)
			if err != nil {
				log.Fatal(err)
			}

			roomsToFiles[c.RoomName] = f
		} else {
			// add user to room
			roomsToUserInfos[c.RoomName] = append(roomsToUserInfos[c.RoomName], userInfo{c.UserName, c.Conn})
		}

		str := fmt.Sprintf("%s@%s> %q joined the room\n", c.UserName, c.RoomName, c.UserName)
		broadcast(str)
		return true
	}

	unregConn := func(str string) {
		userName, roomName, _ := parseMsg(str)
		if (userName != "") && (roomName != "") {
			if _, exists := usrNames[userName]; exists {
				log.Printf("Closing connection with %q", userName)
				broadcast(str)
				if uInfos, exists := roomsToUserInfos[roomName]; exists {
					for idx, ui := range uInfos {
						if ui.userName == userName {
							ui.userConn.Close()
							uInfos = append(uInfos[:idx], uInfos[idx+1:]...)
							break
						}
					}

				}
				delete(usrNames, userName)
			} else {
				log.Printf("Dropped connection with %q", userName)
			}
		}

	}

	defer func() {
		broadcast("admin@global> Server stopping!\n")

		// close existing connections
		for _, uInfos := range roomsToUserInfos {
			for _, ui := range uInfos {
				ui.userConn.Close()
			}
		}

		// close all file handles
		for _, f := range roomsToFiles {
			f.Sync()
			if err := f.Close(); err != nil {
				log.Println(err)
			}
		}
	}()

	for {
		select {
		case c := <-s.register:
			if regConn(c) {
				go c.Readloop()
			}
		case str := <-s.msg:
			broadcast(str)
		case str := <-s.unregister:
			unregConn(str)
		case <-s.stop:
			return
		}
	}
}

var rgxMsg *regexp.Regexp

func init() {
	var err error
	if rgxMsg, err = regexp.Compile(`(?m)(\w+)@(\w+)> (.*$)`); nil != err {
		log.Println("Error compiling the regex")
		os.Exit(1)
	}
}

func parseMsg(inp string) (name, room, msg string) {
	result := rgxMsg.FindStringSubmatch(inp)
	if len(result) != 0 {
		return strings.TrimSpace(result[1]), strings.TrimSpace(result[2]), strings.TrimSpace(result[3])
	}
	return "", "", ""
}
