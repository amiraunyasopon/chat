package client

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"

	"chat/internal/protocol"
)

type Client struct {
	Conn net.Conn
}

func New(conn net.Conn) *Client {
	return &Client{Conn: conn}
}

func (c *Client) Run() {
	defer c.Conn.Close()
	log.Println("Client connected:", c.Conn.RemoteAddr())

	reader := bufio.NewReader(c.Conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("Client disconnected:", c.Conn.RemoteAddr())
			} else {
				log.Printf("Read error: %v", err)
			}
			return
		}

		cmd, ok := protocol.Parse(message)
		if !ok {
			continue
		}

		switch cmd.Name {
		case "PING":
			fmt.Fprintln(c.Conn, "PONG")
		case "ECHO":
			if cmd.Args == "" {
				fmt.Fprintln(c.Conn, "ERR usage: ECHO <text>")
			} else {
				fmt.Fprintln(c.Conn, "OK", cmd.Args)
			}
		case "QUIT":
			fmt.Fprintln(c.Conn, "BYE")
			return
		default:
			fmt.Fprintf(c.Conn, "ERR unknown command %q\n", cmd.Name)
		}
	}
}