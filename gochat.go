package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

)

type Client struct {
	conn *websocket.Conn
	send chan []byte
	name string
}

type Hub struct {
	lock    sync.Mutex
	clients map[*Client]bool
}

func newHub() *Hub {
	return &Hub{clients: make(map[*Client]bool)}
}

func (h *Hub) addClient(c *Client) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.clients[c] = true
}

func (h *Hub) removeClient(c *Client) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.send)
	}
}

func (h *Hub) broadcast(msg []byte, c *Client) {
	timestamp := time.Now().Format(time.RFC3339)
	msg = []byte(fmt.Sprintf("[%s] %s", timestamp, msg))

	h.lock.Lock()
	defer h.lock.Unlock()
	for c := range h.clients {
		c.send <- msg
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for local dev. Lock this down in production.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// readPump reads messages from the browser and broadcasts them
func (c *Client) readPump(h *Hub) {
	defer func() {
		h.removeClient(c)
		c.conn.Close()
		h.broadcast([]byte(c.name+" has left the chat"), c)
	}()

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		full := []byte(c.name + ": " + string(msg))
		log.Println(string(full))
		h.broadcast(full, c)
	}
}

// writePump pushes queued messages out to the browser
func (c *Client) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}

func serveWs(h *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		name = "anonymous"
	}

	client := &Client{conn: conn, send: make(chan []byte, 16), name: name}
	h.addClient(client)
	h.broadcast([]byte(name+" has joined the chat"), client)

	go client.writePump()
	client.readPump(h)
}

func main() {
	h := newHub()

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(h, w, r)
	})

	log.Println("server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}