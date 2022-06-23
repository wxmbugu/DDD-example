package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const version = "1.0.0"

type Server struct {
	Router *mux.Router
	//controllers services.Service
}

func NewServer() *Server {
	mux := mux.NewRouter()

	server := Server{
		Router: mux,
	}
	server.Routes()
	srve := http.Server{
		Addr:        "localhost:9000",
		Handler:     mux,
		ReadTimeout: 10 * time.Second,
	}
	fmt.Println("serving at port :9000")
	srve.ListenAndServe()
	return &server
}

func (server *Server) Routes() {
	server.Router.HandleFunc("/v1/healthcheck", server.Healthcheck).Methods("GET")

}
func (server *Server) Healthcheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "version is %s\n", version)
	fmt.Fprintf(w, "Something works")
}
