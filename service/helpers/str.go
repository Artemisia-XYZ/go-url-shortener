package helpers

import (
	"math/rand"
	"time"
)

var seedRand = rand.New(rand.NewSource(time.Now().UnixNano()))

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func StrRandom(length uint) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seedRand.Int31n(62)]
	}
	return string(b)
}
