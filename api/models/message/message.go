package message

import (
	"time"
)

const (
	MTText    = "chat"
	MTImage   = "image"
	MTGif     = "gif"
	MTSticker = "sticker"
	MTEmoji   = "emoji"
	MTFile    = "file"
)

// I think I don't need id here
type Message struct {
	// ID       string    `json:"id" bson:"_id,omitempty"`
	Type     int       `json:"type" bson:"type"`
	Msg      []byte    `json:"msg" bson:"msg"`
	From     string    `json:"from" bson:"from"` // User ID or email
	TimeSent time.Time `json:"time_sent" bson:"time_sent"`
}
