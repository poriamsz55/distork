package handlers

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/poriamsz55/distork/api/models/message"
	"github.com/poriamsz55/distork/api/models/room"
	"github.com/poriamsz55/distork/api/models/user"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048, // TODO
	WriteBufferSize: 2048, // TODO
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Signal struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

var rooms = make(map[string]*room.Room)
var mu sync.Mutex

func HandeConnection(c echo.Context) error {

	roomsInDB, err := room.GetAllRooms()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Could not get rooms from DB")
	}

	for _, r := range roomsInDB {
		rooms[r.RoomID] = r
	}

	usr := c.Get("user").(*user.User)
	roomID := c.Param("id")

	mu.Lock()
	rm, exists := rooms[roomID]
	if !exists {
		log.Printf("Could not find room")
		return c.String(http.StatusBadRequest, "Room is not exists")
	}
	mu.Unlock()

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("Unable to upgrade http connection %v", err)
		return c.String(http.StatusBadRequest, "Connection failed")
	}
	defer conn.Close()
	log.Println("new client connected")
	broadcastToSelf(rm, usr)

	mu.Lock()
	usr.Conn = conn
	rm.Users = append(rm.Users, usr)
	mu.Unlock()

	// Notify others about the new user
	broadcastToOthers(rm, usr, Signal{Type: "new_user", Data: usr.Username})

	go readMessages(usr, rm)
	return c.String(http.StatusOK, "User disconnected")
}

func readMessages(usr *user.User, rm *room.Room) {
	defer func() {
		mu.Lock()
		defer mu.Unlock()
		for ui, u := range rm.Users {
			if u.Email == usr.Email {
				rm.Users = append(rm.Users[:ui], rm.Users[ui+1:]...)
			}
		}
		usr.Conn.Close()
	}()
	for {
		var signal Signal
		err := usr.Conn.ReadJSON(&signal)

		// if signal.Type == "icecandidate"
		// Send the candidate to all other users in the room
		if err != nil {
			log.Printf("User Disconnected %v-%v", usr, err)
			broadcastToOthers(rm, usr, Signal{Type: "user_left", Data: usr.Username})
			return
		}

		broadcastToOthers(rm, usr, signal)
	}
}

func broadcastToOthers(rm *room.Room, usr *user.User, signal Signal) {
	mu.Lock()
	defer mu.Unlock()

	// TOOD
	if signal.Type == "chat" {
		rm.MessageList = append(rm.MessageList, &message.Message{
			Type:     "chat",
			From:     usr,
			TimeSent: time.Now(),
			Msg:      signal.Data,
		})
	}

	for _, u := range rm.Users {
		if u.Email != usr.Email {
			err := u.Conn.WriteJSON(signal)
			if err != nil {
				log.Printf("error writing messages %v", err)
				return
			}
		}
	}
}

// TODO: load some messages
// by scrolling you can load more
// to show last messages
func broadcastToSelf(rm *room.Room, usr *user.User) {
	mu.Lock()
	defer mu.Unlock()

	for _, data := range rm.MessageList {
		signal := Signal{
			Type: "chat",
			Data: data,
		}
		err := usr.Conn.WriteJSON(signal)
		if err != nil {
			log.Printf("error writing messages %v", err)
			return
		}
	}
}
