package distork

import (
	"encoding/json"
	"sync"

	"github.com/poriamsz55/distork/api/models/message"
	"github.com/poriamsz55/distork/api/models/room"
)

type Hub struct {
	Rooms      map[string]*room.Room `json:"rooms"`
	Register   chan *room.Client
	Unregister chan *room.Client
	Mutex      sync.Mutex
}

func NewHub() *Hub {

	return &Hub{
		Rooms:      make(map[string]*room.Room),
		Register:   make(chan *room.Client),
		Unregister: make(chan *room.Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case registration := <-h.Register:
			h.Mutex.Lock()

			roomsAndUsers := message.Message{
				Type:    "room_list",
				Content: h.Rooms,
				Target:  registration.Username,
			}

			// rooms and users
			rauBytes, _ := json.Marshal(roomsAndUsers)
			registration.Send <- rauBytes
			h.Mutex.Unlock()

		case client := <-h.Unregister:
			h.Mutex.Lock()

			if client.Room != nil {
				if _, ok := client.Room.Clients[client]; ok {
					delete(client.Room.Clients, client)
					close(client.Send)

					usrByte, _ := json.Marshal(client)
					// Notify others about user leaving
					announcement := message.Message{
						Type:    "user_left",
						Content: usrByte,
						RoomId:  client.Room.RoomId,
					}
					announcementBytes, _ := json.Marshal(announcement)
					client.Room.Broadcast <- announcementBytes
				}
			}
			h.Mutex.Unlock()

		}
	}
}
