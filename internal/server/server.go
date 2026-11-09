package server

import (
	"log"
	"net"

	"chat/internal/client"
)

func Run(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Println("Echo server listening on", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting conn:", err)
			continue
		}
		go client.New(conn).Run()
	}
}