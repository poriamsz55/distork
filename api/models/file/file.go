package file

import "time"

// FileInfo struct to hold file name and modification time
type File struct {
	ID       string    `json:"id,omitempty"`
	Filename string    `json:"filename"`
	Size     int64     `json:"size"`
	UserID   string    `json:"user_id"`
	ModTime  time.Time `json:"modTime"`
	Path     string    `json:"path" bson:"path"`
	IsDir    bool      `json:"isDir" bson:"isDir"`
}
