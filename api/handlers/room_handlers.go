package handlers

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/poriamsz55/distork/api/models/message"
	"github.com/poriamsz55/distork/api/models/user"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048, // TODO
	WriteBufferSize: 2048, // TODO
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	Conn *websocket.Conn
	Msg  *message.Message
	Send chan []byte
	User *user.User
}

var clients = make(map[*Client]bool)
var mu sync.Mutex

func HandleWS(c echo.Context) error {

	usr := c.Get("user").(*user.User)

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("Unable to upgrade http connection %v", err)
		return err
	}
	defer conn.Close()
	log.Println("new client connected")

	client := &Client{Conn: conn,
		Send: make(chan []byte),
		User: usr,
		Msg: &message.Message{
			From:     usr.Username,
			TimeSent: time.Now(),
		}}
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
		defer mu.Unlock()
		delete(clients, c)
		c.Conn.Close()
	}()
	for {
		t, msg, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("error reading messages %v", err)
			return
		}

		c.Msg.Type = t
		c.Msg.Msg = msg
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
	defer func() {
		mu.Lock()
		defer mu.Unlock()
		delete(clients, c)
		c.Conn.Close()
	}()
	for msg := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Printf("error writing messages %v", err)
			return
		}
	}
}
