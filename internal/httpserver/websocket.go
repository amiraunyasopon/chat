package httpserver

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

var _ net.Conn = (*webSocketConn)(nil)

type webSocketConn struct {
	conn    *websocket.Conn
	readMu  sync.Mutex
	writeMu sync.Mutex
	buffer  []byte
}

func newWebSocketConn(conn *websocket.Conn) net.Conn {
	return &webSocketConn{conn: conn}
}

func (c *webSocketConn) Read(p []byte) (int, error) {
	c.readMu.Lock()
	defer c.readMu.Unlock()

	if len(c.buffer) == 0 {
		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			return 0, err
		}
		if messageType != websocket.TextMessage {
			return 0, io.ErrUnexpectedEOF
		}
		c.buffer = append(message, '\n')
	}

	n := copy(p, c.buffer)
	c.buffer = c.buffer[n:]
	return n, nil
}

func (c *webSocketConn) Write(p []byte) (int, error) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if err := c.conn.WriteMessage(websocket.TextMessage, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (c *webSocketConn) Close() error {
	return c.conn.Close()
}

func (c *webSocketConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *webSocketConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *webSocketConn) SetDeadline(t time.Time) error {
	c.readMu.Lock()
	defer c.readMu.Unlock()
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if err := c.conn.SetReadDeadline(t); err != nil {
		return err
	}
	return c.conn.SetWriteDeadline(t)
}

func (c *webSocketConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *webSocketConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
