package room

import "github.com/gorilla/websocket"

type Client struct {
	Conn     *websocket.Conn `json:"-" bson:"-"`
	Send     chan []byte     `json:"-" bson:"-"`
	Room     *Room           `json:"room,omitempty" bson:"-"`
	Username string          `json:"username,omitempty" bson:"-"`
}

func NewClient(username string) *Client {
	return &Client{
		Send:     make(chan []byte, 256),
		Username: username,
	}
}
