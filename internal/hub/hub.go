package hub

import (
	"fmt"
	"log"

	"chat/internal/client"
)

type Event interface {
	isEvent()
}

type joinEvent struct {
	c *client.Client
}

type leaveEvent struct {
	c *client.Client
}

type messageEvent struct {
	m client.Message
}

func (joinEvent) isEvent()    {}
func (leaveEvent) isEvent()   {}
func (messageEvent) isEvent() {}

type Hub struct {
	clients map[*client.Client]bool
	events  chan Event
}

func New() *Hub {
	return &Hub{
		clients: make(map[*client.Client]bool),
		events:  make(chan Event, 16),
	}
}

func (h *Hub) Register(c *client.Client)   { h.events <- joinEvent{c: c} }
func (h *Hub) Unregister(c *client.Client) { h.events <- leaveEvent{c: c} }
func (h *Hub) Broadcast(m client.Message)  { h.events <- messageEvent{m: m} }

func (h *Hub) Run() {
	for e := range h.events {
		switch ev := e.(type) {
		case joinEvent:
			h.clients[ev.c] = true
			line := fmt.Sprintf("* %s joined", ev.c.Conn.RemoteAddr())
			h.sendToAllExcept(ev.c, line)
			log.Printf("Hub: client joined (%d connected)", len(h.clients))
		case leaveEvent:
			if h.clients[ev.c] {
				line := fmt.Sprintf("* %s left", ev.c.Conn.RemoteAddr())
				delete(h.clients, ev.c)
				ev.c.Shutdown()
				h.sendToAllExcept(ev.c, line)
				log.Printf("Hub: client left (%d connected)", len(h.clients))
			}
		case messageEvent:
			line := fmt.Sprintf("%s: %s", ev.m.From.Conn.RemoteAddr(), ev.m.Text)
			h.sendToAllExcept(ev.m.From, line)
		default:
			log.Printf("Hub: unknown event type %T", e)
		}
	}
}

func (h *Hub) sendToAllExcept(skip *client.Client, line string) {
	for c := range h.clients {
		if c == skip {
			continue
		}
		c.Send(line)
	}
}
