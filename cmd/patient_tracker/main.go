package main

import (
	_ "github.com/lib/pq"
	"github.com/patienttracker/internal/api"
)

//	"flag"

//const version = "1.0.0"

//Initialize postgres db connection

//const (
//	host     = "localhost"
//	port     = 5432
//	user     = "postgres"
//	password = "secret"
//	dbname   = "patient_tracker"
//)

/*
type r struct {
	service models.AppointmentRepository
}
*/

func main() {
	//flag.IntVar(&config.port, "server port", 3200, "port for server to listen to ...")
	//flag.StringVar(&config.env, "env", "development", "Environment (development|staging|production)")
	//flag.Parse()
	//Initialize logger
	api.NewServer()

	//db.Newdb("postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable")

}

/*
func (service *r) something() {
	service.service.Create()
}
*/
