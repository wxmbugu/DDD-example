package main

import (
	"context"
	"database/sql"
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

func main() {
	var wait time.Duration
	conn := SetupDb("postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable") //TODO: write the database into an env file
	services := services.NewService(conn)
	mux := mux.NewRouter()
	mux.PathPrefix("/debug/").Handler(http.DefaultServeMux)
	server := api.NewServer(services, mux)
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
	go func() {
		ticker := time.NewTicker(20000 * time.Millisecond)
		for range ticker.C {
			server.AppointmentsEmailSender()
		}
	}()
	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srve.ListenAndServe(); err != nil {
			server.Log.Fatal(err)
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
	// closing our channel to stop executing tasks
	server.Log.Info("completing background tasks...")
	server.WaitGroup.Wait()
	server.Worker.Stop()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srve.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	os.Exit(0)
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
