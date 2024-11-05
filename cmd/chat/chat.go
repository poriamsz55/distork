package main

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	Conn *websocket.Conn
	Send chan []byte
}

var clients = make(map[*Client]bool)
var mu sync.Mutex

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/static", "static")

	// Serve HTML page
	e.GET("/", func(c echo.Context) error {
		return c.File("templates/index.html")
	})

	// WebSocket endpoint
	e.GET("/ws", handleConnectionsChat)

	// Start the server
	// e.Logger.Fatal(e.StartTLS(":8443", "localhost+2.pem", "localhost+2-key.pem"))
	e.Logger.Fatal(e.Start(":8080"))
}

func handleConnectionsChat(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := &Client{Conn: conn, Send: make(chan []byte)}
	mu.Lock()
	clients[client] = true
	mu.Unlock()

	go client.readMessages()
	client.writeMessages()
	return nil
}

func (c *Client) readMessages() {
	defer func() {
		mu.Lock()
		delete(clients, c)
		mu.Unlock()
		c.Conn.Close()
	}()
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}
		broadcast(msg)
	}
}

func broadcast(msg []byte) {
	mu.Lock()
	defer mu.Unlock()
	for client := range clients {
		select {
		case client.Send <- msg:
		default:
			close(client.Send)
			delete(clients, client)
		}
	}
}

func (c *Client) writeMessages() {
	for msg := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}
