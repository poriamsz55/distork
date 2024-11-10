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
	ReadBufferSize:  2048, // TODO
	WriteBufferSize: 2048, // TODO
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

	go write(client)
	go readPump(client, hub)

	return c.String(http.StatusOK, "User disconnected")
}

func write(client *room.Client) {
	defer client.Conn.Close()

	for {
		message, ok := <-client.Send
		if !ok {
			client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
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
			break
		}

		var msg message.Message
		if err := json.Unmarshal(msgRcv, &msg); err != nil {
			continue
		}

		// Handle different message types
		switch msg.Type {
		case "create_room":
			hub.Mutex.Lock()
			defer hub.Mutex.Unlock()
			rmDB, err := room.GetRoomByName(msg.RoomId) // here is the room name
			if err != nil {
				errorMsg := message.Message{
					Type:    "error",
					Content: map[string]interface{}{"error": "room is fucked up"},
					Target:  client.Username,
				}

				errMsgByte, _ := json.Marshal(errorMsg)
				client.Send <- errMsgByte
				continue
			}
			_, exists := hub.Rooms[rmDB.RoomId]
			if exists {
				errorMsg := message.Message{
					Type:    "error",
					Content: map[string]interface{}{"error": "room exists"},
					Target:  client.Username,
				}

				errMsgByte, _ := json.Marshal(errorMsg)
				client.Send <- errMsgByte
				continue
			}

			rm := room.NewRoom(msg.RoomId)
			err = rm.AddRoomToDB()
			if err != nil {
				errorMsg := message.Message{
					Type:    "error",
					Content: map[string]interface{}{"error": "room is fucked up in DB"},
					Target:  client.Username,
				}

				errMsgByte, _ := json.Marshal(errorMsg)
				client.Send <- errMsgByte
				continue
			}

			hub.Rooms[rm.RoomId] = rm
			// Notify others in the room about the new user
			announcement := message.Message{
				Type:   "room_created",
				RoomId: rm.RoomId,
			}
			announcementBytes, _ := json.Marshal(announcement)
			client.Send <- announcementBytes

		case "user_joined":
			hub.Mutex.Lock()
			rm, exists := hub.Rooms[msg.RoomId]
			if !exists {
				errorMsg := message.Message{
					Type:    "error",
					Content: map[string]interface{}{"error": "room is not exists"},
					Target:  client.Username,
				}

				errMsgByte, _ := json.Marshal(errorMsg)
				client.Send <- errMsgByte
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
				Content: client.Room,
				Target:  client.Username,
			}
			userListBytes, _ := json.Marshal(userListMsg)
			client.Send <- userListBytes

			hub.Mutex.Unlock()

		case "chat":
			if client.Room != nil {
				client.Room.Broadcast <- msgRcv
			}
		case "signal":
			if client.Room != nil {
				client.Room.Broadcast <- msgRcv
				// Forward WebRTC signaling messages to the target user
				client.Room.Mutex.Lock()
				for client := range client.Room.Clients {
					if client.Username == msg.Target {
						client.Send <- msgRcv
						break
					}
				}
				client.Room.Mutex.Unlock()
			}
		}
	}
}
