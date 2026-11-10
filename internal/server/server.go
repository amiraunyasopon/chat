package server

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"

	"chat/internal/client"
	"chat/internal/httpserver"
	"chat/internal/hub"
)

func Run() error {
	chatPort := os.Getenv("CHAT_PORT")
	if chatPort == "" {
		chatPort = "8080"
	}
	if len(os.Args) > 1 {
		chatPort = os.Args[1]
	}

	wsPort := os.Getenv("WS_PORT")
	if wsPort == "" {
		wsPort = "8081"
	}
	if len(os.Args) > 2 {
		wsPort = os.Args[2]
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

	listener, err := tls.Listen("tcp", ":"+chatPort, config)
	if err != nil {
		return err
	}
	defer listener.Close()

	h := hub.New()
	go h.Run()

	go func() {
		log.Println("HTTP server listening on", ":"+wsPort)
		if err := http.ListenAndServe(":"+wsPort, httpserver.New(h)); err != nil {
			log.Println("HTTP server stopped:", err)
		}
	}()

	log.Println("Chat server listening on", ":"+chatPort)

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
