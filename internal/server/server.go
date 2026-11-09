package server

import (
	"log"
	"net"

	"chat/internal/client"
	"chat/internal/hub"
)

func Run(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	h := hub.New()
	go h.Run()

	log.Println("Chat server listening on", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting conn:", err)
			continue
		}
		cl := client.New(conn, h.Unregister)
		h.Register(cl)
		go cl.WriteLoop()
		go cl.Run()
	}
}
