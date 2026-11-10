package server

import (
	"crypto/tls"
	"log"
	"os"

	"chat/internal/client"
	"chat/internal/hub"
)

func Run() error {
	port := os.Getenv("CHAT_PORT")
	if port == "" {
		port = "8080"
	}
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	certFile := os.Getenv("TLS_CERT_FILE")
	if certFile == "" {
		certFile = "server.crt"
	}
	keyFile := os.Getenv("TLS_KEY_FILE")
	if keyFile == "" {
		keyFile = "server.key"
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	listener, err := tls.Listen("tcp", ":"+port, config)
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
		cl := client.New(conn, h)
		h.Register(cl)
		go cl.WriteLoop()
		go cl.Run()
	}
}
