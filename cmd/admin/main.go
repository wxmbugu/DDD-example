package main

import (
	"database/sql"
	// "fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/patienttracker/internal/services"
)

// TODO: Setup config file
func main() {
	conn := SetupDb("postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable")
	services := services.NewService(conn)
	// var email string

	admin, _ := services.CreateAdmin("mfdoom@mail.com", "mfdoom")
	log.Println(admin)
}

func SetupDb(conn string) *sql.DB {
	db, err := sql.Open("postgres", conn)
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(30)
	db.SetMaxIdleConns(30)
	db.SetConnMaxLifetime(time.Hour)
	return db
}
