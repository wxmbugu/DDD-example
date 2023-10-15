package utils

import (
	"fmt"
	"math/rand"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

var contactRunes = []rune("1234567890")
var englishFirstNames = []string{
	"John", "Jane", "Michael", "Emily", "William", "Olivia", "James", "Emma", "Robert", "Ava",
	"David", "Sophia", "Joseph", "Isabella", "Daniel", "Mia", "Richard", "Charlotte", "Thomas", "Amelia",
}

var englishLastNames = []string{
	"Smith", "Johnson", "Williams", "Brown", "Jones", "Miller", "Davis", "Garcia", "Rodriguez", "Martinez",
	"Wilson", "Anderson", "Taylor", "Thomas", "Moore", "Jackson", "White", "Harris", "Martin", "Thompson",
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Randid(min, max int) int {
	return rand.Intn(max-min) + min
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
func Randfullname() string {
	firstName := englishFirstNames[rand.Intn(len(englishFirstNames))]
	lastName := englishLastNames[rand.Intn(len(englishLastNames))]
	return firstName + " " + lastName
}
func RandEmail(n int) string {
	return fmt.Sprintf("%s@gmail.com", RandString(n))
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
