package hub

import (
	"fmt"
	"log"
	"strings"

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

type nickEvent struct {
	n client.Nickname
}

type directMessageEvent struct {
	d client.DirectMessage
}

func (joinEvent) isEvent()          {}
func (leaveEvent) isEvent()         {}
func (messageEvent) isEvent()       {}
func (nickEvent) isEvent()          {}
func (directMessageEvent) isEvent() {}

type Hub struct {
	clients    map[*client.Client]string
	nicks      map[string]*client.Client
	guestCount int
	events     chan Event
}

func New() *Hub {
	return &Hub{
		clients:    make(map[*client.Client]string),
		nicks:      make(map[string]*client.Client),
		guestCount: 0,
		events:     make(chan Event, 16),
	}
}

func (h *Hub) Register(c *client.Client)            { h.events <- joinEvent{c: c} }
func (h *Hub) Unregister(c *client.Client)          { h.events <- leaveEvent{c: c} }
func (h *Hub) Broadcast(m client.Message)           { h.events <- messageEvent{m: m} }
func (h *Hub) Nickname(n client.Nickname)           { h.events <- nickEvent{n: n} }
func (h *Hub) DirectMessage(d client.DirectMessage) { h.events <- directMessageEvent{d: d} }

func (h *Hub) Run() {
	for e := range h.events {
		switch ev := e.(type) {
		case joinEvent:
			name := h.newGuestName()
			h.clients[ev.c] = name
			h.nicks[name] = ev.c
			ev.c.Send(fmt.Sprintf("* Welcome, %s. Use NICK <name> to change it.", name))
			line := fmt.Sprintf("* %s joined", name)
			h.sendToAllExcept(ev.c, line)
			log.Printf("Hub: client joined (%d connected)", len(h.clients))
		case leaveEvent:
			if name, ok := h.clients[ev.c]; ok {
				line := fmt.Sprintf("* %s left", name)
				delete(h.clients, ev.c)
				ev.c.Shutdown()
				h.sendToAllExcept(ev.c, line)
				log.Printf("Hub: client left (%d connected)", len(h.clients))
			}
		case messageEvent:
			if name, ok := h.clients[ev.m.From]; ok {
				line := fmt.Sprintf("%s: %s", name, ev.m.Text)
				h.sendToAllExcept(ev.m.From, line)
			}
		case nickEvent:
			name := ev.n.Nickname
			if !h.nameTaken(name, ev.n.User) {
				delete(h.nicks, h.clients[ev.n.User])
				h.clients[ev.n.User] = name
				h.nicks[name] = ev.n.User
			}
		case directMessageEvent:
			if name, ok := h.clients[ev.d.From]; ok {
				line := fmt.Sprintf("%s: %s", name, ev.d.Text)
				h.nicks[ev.d.To].Send(line)
			}
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

func (h *Hub) nameTaken(name string, except *client.Client) bool {
	for c, n := range h.clients {
		if c != except && strings.EqualFold(n, name) {
			return true
		}
	}
	return false
}

func (h *Hub) newGuestName() string {
	for {
		h.guestCount++
		name := fmt.Sprintf("guest%d", h.guestCount)
		if !h.nameTaken(name, nil) {
			return name
		}
	}
}
