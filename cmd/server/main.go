package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func handleConn(conn net.Conn) {
	defer conn.Close()
	log.Println("Client connected:", conn.RemoteAddr())

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("Client disconnected:", conn.RemoteAddr())
			} else {
				log.Printf("Read error: %v", err)
			}
			return
		}

		line := strings.TrimSpace(message)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		cmd := strings.ToUpper(parts[0])
		args := ""
		if len(parts) > 1 {
			args = parts[1]
		}

		switch cmd {
		case "PING":
			fmt.Fprintln(conn, "PONG")
		case "ECHO":
			if args == "" {
				fmt.Fprintln(conn, "ERR usage: ECHO <text>")
			} else {
				fmt.Fprintln(conn, "OK", args)
			}
		case "QUIT":
			fmt.Fprintln(conn, "BYE")
			return
		default:
			fmt.Fprintf(conn, "ERR unknown command %q\n", cmd)
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8090")
	if err != nil {
		log.Fatal("Error listening:", err)
	}
	defer listener.Close()

	log.Println("Echo server listening on :8090")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting conn:", err)
			continue
		}
		go handleConn(conn)
	}

}
