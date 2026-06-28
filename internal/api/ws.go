package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yusnelgg/agent-bridge/internal/protocol"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WSClient struct {
	conn *websocket.Conn
	send chan []byte
	done chan struct{}
}

type WSHub struct {
	mu      sync.RWMutex
	clients map[*WSClient]bool
}

func NewWSHub() *WSHub {
	return &WSHub{clients: make(map[*WSClient]bool)}
}

func (h *WSHub) Broadcast(msg *protocol.Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[ws] error marshaling: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.send <- data:
		default:
			log.Printf("[ws] cliente lento, desconectando")
			close(client.done)
			delete(h.clients, client)
		}
	}
}

func (h *WSHub) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[ws] upgrade error: %v", err)
		return
	}

	client := &WSClient{
		conn: conn,
		send: make(chan []byte, 64),
		done: make(chan struct{}),
	}

	h.mu.Lock()
	h.clients[client] = true
	h.mu.Unlock()

	log.Printf("[ws] cliente conectado (%d total)", len(h.clients))

	go client.writePump()
	go client.readPump(h)
}

func (c *WSClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("[ws] write error: %v", err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.done:
			return
		}
	}
}

func (c *WSClient) readPump(hub *WSHub) {
	defer func() {
		hub.mu.Lock()
		delete(hub.clients, c)
		hub.mu.Unlock()
		c.conn.Close()
	}()

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
