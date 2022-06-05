package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	//	"github.com/patienttracker/internal/controllers"
	//"github.com/patienttracker/internal/models"
)

//postgres database connection setup
type Database struct {
	Conn *sql.DB
	Dsn  string
}

func New() (Database, error) {
	conn, err := sql.Open("postgres", "postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	return Database{
		Conn: conn,
	}, nil
}

type Db struct {
	Dns string
}

func Newdb(dns string) Db {
	return Db{
		Dns: dns,
	}
}
