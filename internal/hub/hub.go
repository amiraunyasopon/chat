package hub

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"unicode"
	"unicode/utf8"

	"chat/internal/client"
)

const defaultRoomName = "lobby"
const maxRoomNameLen = 16

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

type joinRoomEvent struct {
	s client.Switch
}

func (joinEvent) isEvent()          {}
func (leaveEvent) isEvent()         {}
func (messageEvent) isEvent()       {}
func (nickEvent) isEvent()          {}
func (directMessageEvent) isEvent() {}
func (joinRoomEvent) isEvent()      {}

type Hub struct {
	clients    map[*client.Client]string // client : nick
	nicks      map[string]*client.Client // nick : client
	guestCount int
	events     chan Event
	rooms      map[string]*Room         // roomName : Room
	clientRoom map[*client.Client]*Room // client : Room
}

func New() *Hub {
	h := &Hub{
		clients:    make(map[*client.Client]string),
		nicks:      make(map[string]*client.Client),
		guestCount: 0,
		events:     make(chan Event, 16),
		rooms:      make(map[string]*Room),
		clientRoom: make(map[*client.Client]*Room),
	}
	h.rooms[defaultRoomName] = newRoom(defaultRoomName)
	return h
}

type Room struct {
	name    string
	clients map[*client.Client]bool // client : bool
}

func newRoom(name string) *Room {
	return &Room{
		name:    name,
		clients: make(map[*client.Client]bool),
	}
}

func (h *Hub) Register(c *client.Client)            { h.events <- joinEvent{c: c} }
func (h *Hub) Unregister(c *client.Client)          { h.events <- leaveEvent{c: c} }
func (h *Hub) Broadcast(m client.Message)           { h.events <- messageEvent{m: m} }
func (h *Hub) Nickname(n client.Nickname)           { h.events <- nickEvent{n: n} }
func (h *Hub) DirectMessage(d client.DirectMessage) { h.events <- directMessageEvent{d: d} }
func (h *Hub) JoinRoom(s client.Switch)             { h.events <- joinRoomEvent{s: s} }

func (h *Hub) Run() {
	for e := range h.events {
		switch ev := e.(type) {
		case joinEvent:
			name := h.newGuestName()
			h.clients[ev.c] = name
			h.nicks[name] = ev.c
			room := h.rooms[defaultRoomName]
			room.clients[ev.c] = true
			h.clientRoom[ev.c] = room
			ev.c.Send(fmt.Sprintf("* Welcome, %s. Use NICK <name> to change it.", name))
			line := fmt.Sprintf("* %s joined", name)
			h.sendToAllExcept(room, ev.c, line)
			log.Printf("Hub: client joined (%d connected)", len(h.clients))
		case leaveEvent:
			if name, ok := h.clients[ev.c]; ok {
				line := fmt.Sprintf("* %s left", name)
				room := h.clientRoom[ev.c]
				delete(h.nicks, h.clients[ev.c])
				delete(h.clients, ev.c)
				h.removeFromRoom(ev.c, room)
				h.sendToAllExcept(room, ev.c, line)
				ev.c.Shutdown()
				log.Printf("Hub: client left (%d connected)", len(h.clients))
			}
		case messageEvent:
			if name, ok := h.clients[ev.m.From]; ok {
				room := h.clientRoom[ev.m.From]
				line := fmt.Sprintf("%s: %s", name, ev.m.Text)
				h.sendToAllExcept(room, ev.m.From, line)
			}
		case nickEvent:
			name := ev.n.Nickname
			if !h.nameTaken(name, ev.n.User) {
				delete(h.nicks, h.clients[ev.n.User])
				h.clients[ev.n.User] = name
				h.nicks[name] = ev.n.User
				ev.n.User.Send(fmt.Sprintf("OK nickname is now %s", name))
			} else {
				ev.n.User.Send(fmt.Sprintf("ERR %s is taken", name))
			}
		case directMessageEvent:
			if name, ok := h.clients[ev.d.From]; ok {
				line := fmt.Sprintf("%s (dm): %s", name, ev.d.Text)
				recipient, ok := h.findNick(ev.d.To)
				if !ok {
					ev.d.From.Send("ERR unknown recipient")
					continue
				}
				recipient.Send(line)
			}
		case joinRoomEvent:
			if name, ok := h.clients[ev.s.User]; ok {
				if err := validateRoomName(ev.s.Room); err != nil {
					ev.s.User.Send("ERR " + err.Error())
					continue
				}
				room, roomFound := h.findRoom(ev.s.Room)
				if roomFound && room == h.clientRoom[ev.s.User] {
					continue
				}
				if !roomFound {
					room = newRoom(ev.s.Room)
					h.rooms[strings.ToLower(ev.s.Room)] = room
				}
				oldRoom := h.clientRoom[ev.s.User]
				h.removeFromRoom(ev.s.User, oldRoom)
				left := fmt.Sprintf("* %s left", name)
				h.sendToAllExcept(oldRoom, ev.s.User, left)
				room.clients[ev.s.User] = true
				h.clientRoom[ev.s.User] = room
				join := fmt.Sprintf("* %s joined", name)
				h.sendToAllExcept(room, ev.s.User, join)
				ev.s.User.Send(fmt.Sprintf("OK %s joined", room.name))
			}
		default:
			log.Printf("Hub: unknown event type %T", e)
		}
	}
}

func (h *Hub) sendToAllExcept(room *Room, skip *client.Client, line string) {
	for c := range room.clients {
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

func (h *Hub) findNick(name string) (*client.Client, bool) {
	for nick, c := range h.nicks {
		if strings.EqualFold(nick, name) {
			return c, true
		}
	}
	return nil, false
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

func (h *Hub) findRoom(name string) (*Room, bool) {
	room, ok := h.rooms[strings.ToLower(name)]
	if ok {
		return room, true
	}
	return nil, false
}

func validateRoomName(name string) error {
	if name == "" {
		return errors.New("room name cannot be empty")
	}
	if strings.IndexFunc(name, unicode.IsSpace) >= 0 {
		return errors.New("room name cannot contain spaces")
	}
	if utf8.RuneCountInString(name) > maxRoomNameLen {
		return fmt.Errorf("room name is too long (max %d)", maxRoomNameLen)
	}
	return nil
}

func (h *Hub) removeFromRoom(c *client.Client, room *Room) {
	delete(room.clients, c)
	delete(h.clientRoom, c)
	if len(room.clients) == 0 && room.name != defaultRoomName {
		delete(h.rooms, strings.ToLower(room.name))
	}
}
