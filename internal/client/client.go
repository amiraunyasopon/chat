package client

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"

	"chat/internal/protocol"
)

const outBuffer = 16

type Message struct {
	From *Client
	Text string
}

type Nickname struct {
	User     *Client
	Nickname string
}

type DirectMessage struct {
	From *Client
	To   string
	Text string
}

type Switch struct {
	User *Client
	Room string
}

type Handler interface {
	Unregister(*Client)
	Broadcast(Message)
	Nickname(Nickname)
	DirectMessage(DirectMessage)
	JoinRoom(Switch)
}

type Client struct {
	Conn    net.Conn
	out     chan string
	handler Handler
}

func New(conn net.Conn, handler Handler) *Client {
	return &Client{
		Conn:    conn,
		out:     make(chan string, outBuffer),
		handler: handler,
	}
}

func (c *Client) Send(line string) bool {
	select {
	case c.out <- line:
		return true
	default:
		log.Printf("Client %s too slow, disconnecting", c.Conn.RemoteAddr())
		c.Conn.Close()
		return false
	}
}

func (c *Client) Shutdown() {
	close(c.out)
}

func (c *Client) WriteLoop() {
	defer c.Conn.Close()
	for line := range c.out {
		if _, err := fmt.Fprintln(c.Conn, line); err != nil {
			log.Printf("Write error: %v", err)
			return
		}
	}
}

func (c *Client) Run() {
	defer c.handler.Unregister(c)
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

		cmd, ok := protocol.ParseCommand(message)
		if !ok {
			continue
		}

		switch cmd.Name {
		case "PING":
			c.Send("PONG")
		case "ECHO":
			if cmd.Args == "" {
				c.Send("ERR usage: ECHO <text>")
			} else {
				c.Send("OK " + cmd.Args)
			}
		case "MSG":
			if cmd.Args == "" {
				c.Send("ERR usage: MSG <text>")
			} else {
				c.handler.Broadcast(Message{From: c, Text: cmd.Args})
			}
		case "NICK":
			if cmd.Args == "" {
				c.Send("ERR usage: NICK <text>")
			} else {
				c.handler.Nickname(Nickname{User: c, Nickname: cmd.Args})
			}
		case "DM":
			if cmd.Args == "" {
				c.Send("ERR usage: DM <user> <text>")
			} else {
				if dm, ok := protocol.ParseDM(cmd.Args); ok {
					c.handler.DirectMessage(DirectMessage{From: c, To: dm.To, Text: dm.Message})
				} else {
					c.Send("ERR usage: DM <user> <text>")
				}
			}
		case "JOIN":
			if cmd.Args == "" {
				c.Send("ERR usage: JOIN <room>")
			} else {
				c.handler.JoinRoom(Switch{User: c, Room: cmd.Args})
			}
		case "QUIT":
			c.Send("BYE")
			return
		default:
			c.Send(fmt.Sprintf("ERR unknown command %q", cmd.Name))
		}
	}
}
