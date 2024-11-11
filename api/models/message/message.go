package message

import (
	"time"
)

type Message struct {
	Type     string      `json:"type" bson:"type"`
	RoomId   string      `json:"room_id,omitempty" bson:"room_id,omitempty"`
	From     string      `json:"from,omitempty" bson:"from,omitempty"` // client username
	Content  interface{} `json:"content,omitempty" bson:"content,omitempty"`
	Target   string      `json:"target,omitempty" bson:"target,omitempty"`
	TimeSent time.Time   `json:"time_sent,omitempty" bson:"time_sent,omitempty"`
}
