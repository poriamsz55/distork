package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/poriamsz55/distork/api/models/distork"
	"github.com/poriamsz55/distork/api/models/message"
	"github.com/poriamsz55/distork/api/models/room"
	"github.com/poriamsz55/distork/api/models/user"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024, // TODO
	WriteBufferSize: 1024, // TODO
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleConnection(c echo.Context, hub *distork.Hub) error {

	usr := c.Get("user").(*user.User)

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("Unable to upgrade http connection %v", err)
		return c.String(http.StatusBadRequest, "Connection failed")
	}
	defer conn.Close()
	log.Println("new client connected")

	client := room.NewClient(usr.Username)
	client.Conn = conn

	hub.Register <- client

	go readPump(client, hub)
	write(client)

	return c.String(http.StatusOK, "User disconnected")
}

func write(client *room.Client) {
	defer client.Conn.Close()

	for {
		message, ok := <-client.Send
		if !ok {
			client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			continue
		}

		if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}

func readPump(client *room.Client, hub *distork.Hub) {
	defer func() {
		hub.Unregister <- client
		client.Conn.Close()
	}()

	for {
		_, msgRcv, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg message.Message
		if err := json.Unmarshal(msgRcv, &msg); err != nil {
			log.Printf("error unmarshaling message: %v", err)
			errorMsg := message.Message{
				Type:    "error",
				Content: map[string]interface{}{"error": "Invalid message format"},
			}
			errorBytes, _ := json.Marshal(errorMsg)
			client.Send <- errorBytes
			continue
		}

		// Handle different message types
		switch msg.Type {
		case "create_room":
			hub.Mutex.Lock()
			_, err := room.GetRoomByNameDB(msg.RoomId) // here is the room name
			if err == nil {
				errorMsg := message.Message{
					Type:    "error",
					Content: map[string]interface{}{"error": "room is fucked up"},
				}

				errMsgByte, _ := json.Marshal(errorMsg)
				client.Send <- errMsgByte
				hub.Mutex.Unlock()
				continue
			}
			_, exists := hub.Rooms[msg.RoomId]
			if exists {
				errorMsg := message.Message{
					Type:    "error",
					Content: map[string]interface{}{"error": "room exists"},
				}

				errMsgByte, _ := json.Marshal(errorMsg)
				client.Send <- errMsgByte
				hub.Mutex.Unlock()
				continue
			}

			rm := room.NewRoom(msg.RoomId)
			err = rm.AddRoomToDB()
			if err != nil {
				errorMsg := message.Message{
					Type:    "error",
					Content: map[string]interface{}{"error": "room is fucked up in DB"},
				}

				errMsgByte, _ := json.Marshal(errorMsg)
				client.Send <- errMsgByte
				hub.Mutex.Unlock()
				continue
			}

			go rm.Run()
			hub.Rooms[rm.RoomId] = rm
			// Notify others in the room about the new user
			announcement := message.Message{
				Type:    "room_created",
				Content: rm,
			}
			announcementBytes, _ := json.Marshal(announcement)
			client.Send <- announcementBytes
			hub.Mutex.Unlock()

		case "join_room":
			hub.Mutex.Lock()
			rm, exists := hub.Rooms[msg.RoomId]
			if !exists {
				errorMsg := message.Message{
					Type:    "error",
					Content: map[string]interface{}{"error": "room is not exists"},
				}

				errMsgByte, _ := json.Marshal(errorMsg)
				client.Send <- errMsgByte
				hub.Mutex.Unlock()
				continue
			}

			client.Room = rm
			rm.Clients[client] = true

			// Notify others in the room about the new user
			announcement := message.Message{
				Type:   "user_joined",
				From:   client.Username,
				RoomId: msg.RoomId,
			}
			announcementBytes, _ := json.Marshal(announcement)
			rm.Broadcast <- announcementBytes

			// Send current users list to the new client
			userList := []string{}
			for cl := range rm.Clients {
				if cl.Username != client.Username {
					userList = append(userList, client.Username)
				}
			}
			userListMsg := message.Message{
				Type:    "user_list",
				Content: userList,
			}
			userListBytes, _ := json.Marshal(userListMsg)
			client.Send <- userListBytes

			hub.Mutex.Unlock()

		case "chat":
			if client.Room != nil {
				client.Room.Broadcast <- msgRcv
			}
		case "signal":
			log.Printf("Signaling message from %s to %s of type %s",
				msg.From, msg.Target, msg.Signal)
			client.Room.Mutex.Lock()
			targetFound := false
			for cl := range client.Room.Clients {
				if cl.Username == msg.Target {
					cl.Send <- msgRcv
					targetFound = true
					break
				}
			}
			if !targetFound {
				log.Printf("Target user %s not found in room", msg.Target)
				errorMsg := message.Message{
					Type:    "error",
					Content: map[string]interface{}{"error": "Target user not found"},
				}
				errorBytes, _ := json.Marshal(errorMsg)
				client.Send <- errorBytes
			}
			client.Room.Mutex.Unlock()
		}
	}
}
