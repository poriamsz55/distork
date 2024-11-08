package main

import (
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/exp/rand"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type Client struct {
	conn     *websocket.Conn
	send     chan []byte
	username string
}

type Message struct {
	Type    string `json:"type"`
	From    string `json:"from"`
	To      string `json:"to,omitempty"`
	Content string `json:"content,omitempty"`
	Signal  string `json:"signal,omitempty"`
}

type ChatHub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.Mutex
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func newHub() *ChatHub {
	return &ChatHub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *ChatHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			// Notify others about new user
			userList := []string{}
			for c := range h.clients {
				userList = append(userList, c.username)
			}
			userListMsg, _ := json.Marshal(Message{
				Type:    "users",
				Content: strings.Join(userList, ","),
			})
			for c := range h.clients {
				c.send <- userListMsg
			}
			h.mutex.Unlock()

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				// Notify others about user leaving
				userList := []string{}
				for c := range h.clients {
					userList = append(userList, c.username)
				}
				userListMsg, _ := json.Marshal(Message{
					Type:    "users",
					Content: strings.Join(userList, ","),
				})
				for c := range h.clients {
					c.send <- userListMsg
				}
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}

			h.mutex.Lock()
			switch msg.Type {
			case "signal":
				// Handle WebRTC signaling
				for client := range h.clients {
					if client.username == msg.To {
						client.send <- message
						break
					}
				}
			default:
				// Broadcast to all clients
				for client := range h.clients {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
			h.mutex.Unlock()
		}
	}
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Initialize template rendering
	t := &Template{
		templates: template.Must(template.ParseGlob("static/*.html")),
	}
	e.Renderer = t

	// Initialize hub
	hub := newHub()
	go hub.run()

	// Routes
	e.Static("/static", "static")
	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", nil)
	})

	e.GET("/ws", func(c echo.Context) error {
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}

		username := c.QueryParam("username")
		if username == "" {
			username = "Anonymous" + strconv.Itoa(rand.Intn(1000))
		}

		client := &Client{
			conn:     ws,
			send:     make(chan []byte, 256),
			username: username,
		}
		hub.register <- client

		go func() {
			defer func() {
				hub.unregister <- client
				client.conn.Close()
			}()

			for {
				_, message, err := client.conn.ReadMessage()
				if err != nil {
					break
				}
				hub.broadcast <- message
			}
		}()

		go func() {
			defer client.conn.Close()
			for {
				message, ok := <-client.send
				if !ok {
					client.conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				err := client.conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					return
				}
			}
		}()

		return nil
	})

	e.Logger.Fatal(e.Start(":8080"))
}
