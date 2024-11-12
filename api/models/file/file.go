package file

import "time"

// FileInfo struct to hold file name and modification time
type File struct {
	Filename  string    `json:"filename" bson:"filename"`
	Size      int64     `json:"size" bson:"size"`
	UUsername string    `json:"u_username" bson:"u_username"`
	ModTime   time.Time `json:"mod_time" bson:"mod_time"`
	Path      string    `json:"path" bson:"path"`
	IsDir     bool      `json:"is_dir" bson:"is_dir"`
}
