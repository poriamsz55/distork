package message

import (
	"time"

	"github.com/poriamsz55/distork/api/models/user"
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
	ID       string      `json:"id" bson:"_id,omitempty"`
	Type     string      `json:"type" bson:"type"`
	Msg      interface{} `json:"msg" bson:"msg"`
	From     *user.User  `json:"from" bson:"from"`
	TimeSent time.Time   `json:"time_sent" bson:"time_sent"`
}
