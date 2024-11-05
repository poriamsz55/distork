package room

import "github.com/poriamsz55/distork/api/models/message"

type Room struct {
	ID          string            `json:"id" bson:"_id,omitempty"`
	RoomID      string            `json:"room_id" bson:"room_id"`
	MessageList []message.Message `json:"message_list" bson:"message_list"`
}

