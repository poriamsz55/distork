package utils

import (
	"time"

	"golang.org/x/exp/rand"
)

func GenerateUUID() string {
	rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789")
	b := make([]rune, 8)

	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}
