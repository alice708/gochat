package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type client struct {
	conn net.Conn
	name string
}

type hub struct {
	lock    sync.Mutex
	clients map[*client]bool
}

func newHub() *hub {
	return &hub{clients: make(map[*client]bool)}
}

func (hub *hub) addClient(c *client) {
	hub.lock.Lock()
	defer hub.lock.Unlock()
	hub.clients[c] = true
}

func (hub *hub) removeClient(c *client) {
	hub.lock.Lock()
	defer hub.lock.Unlock()
	delete(hub.clients, c)
}

func (hub *hub) broadcast(msg string, sender *client) {
	timestamp := time.Now().Format(time.RFC3339)
	msg = fmt.Sprintf("[%s] %s", timestamp, msg)
	hub.lock.Lock()
	defer hub.lock.Unlock()
	for c := range hub.clients {
		fmt.Fprintln(c.conn, msg)
	}
}

func handleConnection(conn net.Conn, hub *hub) {
	defer conn.Close()

	client := &client{conn: conn}

	fmt.Fprint(conn, "Enter your name: ")
	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		return
	}
	client.name = scanner.Text()

	hub.addClient(client)
	hub.broadcast(fmt.Sprintf("* %s has joined the chat *", client.name), client)
	defer func() {
		hub.removeClient(client)
		hub.broadcast(fmt.Sprintf("* %s has left the chat *", client.name), client)
	}()

	for scanner.Scan() {
		text := scanner.Text()
		msg := fmt.Sprintf("%s: %s", client.name, text)
		log.Println(msg)
		hub.broadcast(msg, client)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("connection error for %s: %v", client.name, err)
	}
}

func main() {
	listener, err := net.Listen("tcp", "localhost:9000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()
	log.Println("chat server listening on :9000")

	hub := newHub()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}
		go handleConnection(conn, hub)
	}
}
