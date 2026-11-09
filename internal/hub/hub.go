package hub

import (
	"fmt"
	"log"

	"chat/internal/client"
)

type Hub struct {
	clients    map[*client.Client]bool
	register   chan *client.Client
	unregister chan *client.Client
	broadcast  chan client.Message
}

func New() *Hub {
	return &Hub{
		clients:    make(map[*client.Client]bool),
		register:   make(chan *client.Client),
		unregister: make(chan *client.Client),
		broadcast:  make(chan client.Message, 16),
	}
}

func (h *Hub) Register(c *client.Client)   { h.register <- c }
func (h *Hub) Unregister(c *client.Client) { h.unregister <- c }
func (h *Hub) Broadcast(m client.Message)  { h.broadcast <- m }

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = true
			log.Printf("Hub: client joined (%d connected)", len(h.clients))
		case c := <-h.unregister:
			if h.clients[c] {
				delete(h.clients, c)
				c.Shutdown()
				log.Printf("Hub: client left (%d connected)", len(h.clients))
			}
		case m := <-h.broadcast:
			line := fmt.Sprintf("%s: %s", m.From.Conn.RemoteAddr(), m.Text)
			for c := range h.clients {
				if c == m.From {
					continue
				}
				c.Send(line)
			}
		}
	}
}
