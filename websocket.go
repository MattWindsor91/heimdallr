package main

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Wspool struct {
	broadcast            chan []byte
	register, unregister chan *wsConn
	connections          map[*wsConn]bool
	quit                 bool
	wg                   *sync.WaitGroup
}

func NewWspool(wg *sync.WaitGroup) (wspool *Wspool) {
	wspool = &Wspool{
		broadcast:   make(chan []byte),
		register:    make(chan *wsConn),
		unregister:  make(chan *wsConn),
		connections: make(map[*wsConn]bool),
		wg:          wg,
	}
	return
}

func (wspool *Wspool) closeConn(conn *wsConn) {
	delete(wspool.connections, conn)
	close(conn.send)
}

func (wspool *Wspool) run() {
	wspool.wg.Add(1)
	for {
		select {
		case payload, ok := <-wspool.broadcast:
			if !ok { // channel has been closed, shutdown
				for conn := range wspool.connections {
					wspool.closeConn(conn)
				}
				wspool.quit = true
			}
			for conn := range wspool.connections {
				select {
				case conn.send <- payload:
				default:
					wspool.closeConn(conn)
				}
			}
		case conn := <-wspool.register:
			wspool.connections[conn] = true
		case conn := <-wspool.unregister:
			if _, ok := wspool.connections[conn]; ok {
				wspool.closeConn(conn)
			}
		}
		if wspool.quit {
			wspool.wg.Done()
			break
		}
	}
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// Wraps the websocket conn and a send channel in a handy struct which can
// be passed to the websocket pool
type wsConn struct {
	ws   *websocket.Conn
	send chan []byte
}

// write writes a message with the given message type and payload.
func (c *wsConn) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// writeLoop writes any messages coming down the send channel and pings the
// client every pingPeriod
func (c *wsConn) writeLoop() {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-pingTicker.C:
			if err := c.write(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
