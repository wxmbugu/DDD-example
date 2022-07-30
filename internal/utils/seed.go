package utils

import (
	"fmt"
	"math/rand"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

var contactRunes = []rune("1234567890")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func RandUsername(n int) string {
	return RandString(n)
}
func Randfullname(n int) string {
	return RandString(n) + " " + RandString(n)
}

func RandEmail(n int) string {
	return RandString(n) + fmt.Sprintln("@mail.com")
}

func RandContact(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = contactRunes[rand.Intn(len(contactRunes))]
	}
	return string(b)
}

func Randate() time.Time {
	min := time.Date(1970, 1, 0, 0, 0, 0, 0, time.Local).Unix()
	max := time.Date(2070, 1, 0, 0, 0, 0, 0, time.Local).Unix()
	delta := max - min
	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0)
}
