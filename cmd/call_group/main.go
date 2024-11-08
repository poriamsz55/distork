package main

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Client struct {
	conn     *websocket.Conn
	send     chan []byte
	room     *Room
	username string
}

type Room struct {
	name      string
	clients   map[*Client]bool
	broadcast chan []byte
	mutex     sync.Mutex
}

type ChatHub struct {
	rooms      map[string]*Room
	register   chan *ClientRegistration
	unregister chan *Client
	mutex      sync.Mutex
}

type ClientRegistration struct {
	client   *Client
	room     string
	username string
}

type Message struct {
	Type     string `json:"type"`
	Room     string `json:"room,omitempty"`
	Username string `json:"username,omitempty"`
	Content  string `json:"content,omitempty"`
	Signal   string `json:"signal,omitempty"`
	Target   string `json:"target,omitempty"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func newRoom(name string) *Room {
	return &Room{
		name:      name,
		clients:   make(map[*Client]bool),
		broadcast: make(chan []byte),
	}
}

func newHub() *ChatHub {
	return &ChatHub{
		rooms:      make(map[string]*Room),
		register:   make(chan *ClientRegistration),
		unregister: make(chan *Client),
	}
}

func (h *ChatHub) run() {
	for {
		select {
		case registration := <-h.register:
			h.mutex.Lock()
			room, exists := h.rooms[registration.room]
			if !exists {
				room = newRoom(registration.room)
				h.rooms[registration.room] = room
				go room.run()
			}

			registration.client.room = room
			registration.client.username = registration.username
			room.clients[registration.client] = true

			// Notify others in the room about the new user
			announcement := Message{
				Type:     "user_joined",
				Username: registration.username,
				Room:     registration.room,
			}
			announcementBytes, _ := json.Marshal(announcement)
			room.broadcast <- announcementBytes

			// Send current users list to the new client
			userList := []string{}
			for client := range room.clients {
				if client.username != registration.username {
					userList = append(userList, client.username)
				}
			}
			userListMsg := Message{
				Type:    "user_list",
				Content: registration.room,
				Signal:  "",
				Target:  registration.username,
			}
			userListBytes, _ := json.Marshal(userListMsg)
			registration.client.send <- userListBytes

			h.mutex.Unlock()

		case client := <-h.unregister:
			h.mutex.Lock()
			if client.room != nil {
				if _, ok := client.room.clients[client]; ok {
					delete(client.room.clients, client)
					close(client.send)

					// Notify others about user leaving
					announcement := Message{
						Type:     "user_left",
						Username: client.username,
						Room:     client.room.name,
					}
					announcementBytes, _ := json.Marshal(announcement)
					client.room.broadcast <- announcementBytes
				}
			}
			h.mutex.Unlock()
		}
	}
}

func (r *Room) run() {
	for {
		message := <-r.broadcast
		r.mutex.Lock()
		for client := range r.clients {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(r.clients, client)
			}
		}
		r.mutex.Unlock()
	}
}

func (c *Client) readPump(hub *ChatHub) {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		// Handle different message types
		switch msg.Type {
		case "chat":
			if c.room != nil {
				c.room.broadcast <- message
			}
		case "signal":
			// Forward WebRTC signaling messages to the target user
			c.room.mutex.Lock()
			for client := range c.room.clients {
				if client.username == msg.Target {
					client.send <- message
					break
				}
			}
			c.room.mutex.Unlock()
		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		message, ok := <-c.send
		if !ok {
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}

func main() {
	e := echo.New()
	hub := newHub()
	go hub.run()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/", "static")

	e.GET("/ws", func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}

		client := &Client{
			conn: conn,
			send: make(chan []byte, 256),
		}

		// Wait for initial message with room and username
		_, message, err := conn.ReadMessage()
		if err != nil {
			conn.Close()
			return err
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			conn.Close()
			return err
		}

		registration := &ClientRegistration{
			client:   client,
			room:     msg.Room,
			username: msg.Username,
		}

		hub.register <- registration

		go client.writePump()
		go client.readPump(hub)

		return nil
	})

	e.Logger.Fatal(e.Start(":8080"))
}
