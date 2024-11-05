package room

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/poriamsz55/distork/api/models/message"
	"github.com/poriamsz55/distork/api/models/user"
	config "github.com/poriamsz55/distork/configs"
	"github.com/poriamsz55/distork/database"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/exp/rand"
)

type Room struct {
	ID          string             `json:"id" bson:"_id,omitempty"`
	RoomID      string             `json:"room_id" bson:"room_id"`
	MessageList []*message.Message `json:"message_list" bson:"message_list"`
	Users       []*user.User       `json:"-" bson:"-"`
}

func CreateRoom(c echo.Context) error {

	rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789")
	b := make([]rune, 8)

	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	roomID := string(b)
	rm := &Room{
		RoomID:      roomID,
		MessageList: []*message.Message{},
	}

	err := rm.AddRoomToDB()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, rm)
}

func (r *Room) AddRoomToDB() error {

	collection := database.Collection(config.GetConfigDB().RoomColl)
	// check if exists
	var rm Room
	err := collection.FindOne(context.TODO(), bson.M{"room_id": r.RoomID}).Decode(&rm)
	if err == nil {
		return nil
	}

	// Add room
	_, err = collection.InsertOne(context.TODO(), r)
	if err != nil {
		return err
	}

	return nil
}

func GetRoomByRoomId(roomId string) (*Room, error) {

	collection := database.Collection(config.GetConfigDB().RoomColl)
	// check if exists
	var rm Room
	err := collection.FindOne(context.TODO(), bson.M{"room_id": roomId}).Decode(&rm)
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
