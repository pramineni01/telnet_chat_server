package cmd

import (
	"flag"
	"log"

	"github.com/pramineni01/telnet_chat_server/server"
)

func Execute() {
	log.SetPrefix("chat: ")
	addr := flag.String("addr", "localhost:4000", "listen address")
	flag.Parse()

	log.Fatal(server.ListenAndServe(*addr))
}
