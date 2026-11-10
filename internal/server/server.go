package server

import (
	"log"
	"net"
	"os"

	"chat/internal/client"
	"chat/internal/hub"
)

func Run() error {
	port := "8080"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	defer listener.Close()

	h := hub.New()
	go h.Run()

	log.Println("Chat server listening on", ":"+port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting conn:", err)
			continue
		}
		cl := client.New(conn, h.Broadcast, h.Unregister, h.Nickname, h.DirectMessage, h.JoinRoom)
		h.Register(cl)
		go cl.WriteLoop()
		go cl.Run()
	}
}
