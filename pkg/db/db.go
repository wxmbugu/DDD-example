package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

//postgres database connection setup

type Db struct {
	db *sql.DB
}

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "secret"
	dbname   = "patient_tracker"
)

//Initialize postgres db connection
func Initialize() (*Db, error) {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	dbconn, err := sql.Open("postgres", psqlconn)
	if err != nil {
		log.Fatal(err)
	}
	defer dbconn.Close()

	db := Db{
		db: dbconn,
	}
	return &db, nil
}
