package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	//	"github.com/patienttracker/internal/controllers"
	//"github.com/patienttracker/internal/models"
)

//postgres database connection setup

func New() (*sql.DB, error) {
	conn, err := sql.Open("postgres", "postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	conn.SetMaxOpenConns(95)
	conn.SetMaxIdleConns(5)

	return conn, nil
}

func Ok(o sql.DB) *sql.DB {
	conn, err := sql.Open("postgres", "postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	conn.SetMaxOpenConns(95)
	conn.SetMaxIdleConns(5)

	return &o
}
