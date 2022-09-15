package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/patienttracker/internal/api"
	"github.com/patienttracker/internal/services"
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
// TODO: Enum type for Bloodgroup i.e: A,B,AB,O
// TODO: Password updated at field
// NOTE: Work on cancel appointments and delete appointments
// TODO: Work on Update structs on api calls
// TODO: Authorithation middleware
func main() {
	var wait time.Duration
	//flag.IntVar(&config.port, "server port", 3200, "port for server to listen to ...")
	//flag.StringVar(&config.env, "env", "development", "Environment (development|staging|production)")
	//flag.Parse()
	//Initialize logger
	conn := api.SetupDb("postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable")
	services := services.NewService(conn)
	mux := mux.NewRouter()
	server := api.NewServer(services, mux)
	//	server.Log.PrintInfo("Connected to db successfully")
	srve := http.Server{
		Addr:         "localhost:9000",
		Handler:      server.Router,
		ErrorLog:     log.New(server.Log, "", 0),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	err := conn.Ping()
	if err != nil {
		server.Log.Fatal(err)
	} else {
		server.Log.Info("Connected to db successfully")
	}
	server.Log.Info(fmt.Sprintf("Serving at %s", srve.Addr))
	//srve.ListenAndServe()
	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srve.ListenAndServe(); err != nil {
			server.Log.Debug(err.Error())
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srve.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	os.Exit(0)
}
