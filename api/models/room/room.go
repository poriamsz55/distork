package room

import (
	"context"
	"sync"

	config "github.com/poriamsz55/distork/configs"
	"github.com/poriamsz55/distork/database"
	"github.com/poriamsz55/distork/utils"
	"go.mongodb.org/mongo-driver/bson"
)

type Room struct {
	RoomId    string           `json:"room_id" bson:"room_id"`
	Name      string           `json:"name,omitempty" bson:"name"`
	Clients   map[*Client]bool `json:"clients,omitempty" bson:"-"`
	Broadcast chan []byte      `json:"-" bson:"-"`
	Mutex     sync.Mutex       `json:"-" bson:"-"`
}

func NewRoom(name string) *Room {
	return &Room{
		RoomId:    utils.GenerateUUID(),
		Name:      name,
		Clients:   make(map[*Client]bool),
		Broadcast: make(chan []byte),
	}
}

func (r *Room) AddRoomToDB() error {

	collection := database.Collection(config.GetConfigDB().RoomColl)
	// check if exists
	var rm Room
	err := collection.FindOne(context.Background(), bson.M{"room_id": r.RoomId}).Decode(&rm)
	if err == nil {
		return nil
	}

	// Add room
	_, err = collection.InsertOne(context.Background(), r)
	if err != nil {
		return err
	}

	return nil
}

func GetRoomByRoomId(roomId string) (*Room, error) {

	collection := database.Collection(config.GetConfigDB().RoomColl)
	// check if exists
	var rm Room
	err := collection.FindOne(context.Background(), bson.M{"room_id": roomId}).Decode(&rm)
	if err != nil {
		return nil, err
	}

	return &rm, nil
}

func GetRoomByNameDB(name string) (*Room, error) {

	collection := database.Collection(config.GetConfigDB().RoomColl)
	// check if exists
	var rm Room
	err := collection.FindOne(context.Background(), bson.M{"name": name}).Decode(&rm)
	if err != nil {
		return nil, err
	}

	return &rm, nil
}

func GetAllRooms() ([]*Room, error) {

	collection := database.Collection(config.GetConfigDB().RoomColl)
	// check if exists
	var rooms []Room
	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}

	err = cursor.All(context.Background(), &rooms)
	if err != nil {
		return nil, err
	}

	roomsAddr := []*Room{}
	for ri := range rooms {
		roomsAddr = append(roomsAddr, &rooms[ri])
	}

	return roomsAddr, nil
}

func (r *Room) Run() {
	for {
		message := <-r.Broadcast
		r.Mutex.Lock()
		for client := range r.Clients {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(r.Clients, client)
			}
		}
		r.Mutex.Unlock()
	}
}
