package controllers

import (
	//"context"
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"os"
	"testing"
	//	"github.com/patienttracker/internal/models"
)

var controllers Controllers

func TestMain(m *testing.M) {

	conn, err := sql.Open("postgres", "postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	controllers = New(conn)
	os.Exit(m.Run())
}
