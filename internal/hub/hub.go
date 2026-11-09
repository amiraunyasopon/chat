package hub

import (
	"log"

	"chat/internal/client"
)

type Hub struct {
	clients    map[*client.Client]bool
	register   chan *client.Client
	unregister chan *client.Client
}

func New() *Hub {
	return &Hub{
		clients:    make(map[*client.Client]bool),
		register:   make(chan *client.Client),
		unregister: make(chan *client.Client),
	}
}

func (h *Hub) Register(c *client.Client)   { h.register <- c }
func (h *Hub) Unregister(c *client.Client) { h.unregister <- c }

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
		}
	}
}
