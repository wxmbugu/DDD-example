package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/patienttracker/internal/services"
)

const version = "1.0.0"

type Server struct {
	Router   *mux.Router
	Services services.Service
}

func NewServer() *Server {
	mux := mux.NewRouter()

	server := Server{
		Router: mux,
	}
	server.Routes()
	srve := http.Server{
		Addr:         "localhost:9000",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	fmt.Println("serving at port :9000")
	srve.ListenAndServe()
	return &server
}

func (server *Server) Routes() {
	http.Handle("/", server.Router)
	server.Router.Use(server.contentTypeMiddleware)
	server.Router.HandleFunc("/v1/healthcheck", server.Healthcheck).Methods("GET")
	server.Router.HandleFunc("/v1/department", server.createdepartment).Methods("POST")
	server.Router.HandleFunc("/v1/department/{id:[0-9]+}", server.deletedepartment).Methods("DELETE")
	server.Router.HandleFunc("/v1/departments", server.findalldepartment).Methods("GET")
	server.Router.HandleFunc("/v1/department/{id:[0-9]+}", server.updatedepartment).Methods("POST")
	err := server.Router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			fmt.Println("ROUTE:", pathTemplate)
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			fmt.Println("Path regexp:", pathRegexp)
		}
		queriesTemplates, err := route.GetQueriesTemplates()
		if err == nil {
			fmt.Println("Queries templates:", strings.Join(queriesTemplates, ","))
		}
		queriesRegexps, err := route.GetQueriesRegexp()
		if err == nil {
			fmt.Println("Queries regexps:", strings.Join(queriesRegexps, ","))
		}
		methods, err := route.GetMethods()
		if err == nil {
			fmt.Println("Methods:", strings.Join(methods, ","))
		}
		fmt.Println()
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

}
func (server *Server) Healthcheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "status: available\n")
	fmt.Fprintf(w, "version: %s\n", version)
	fmt.Fprintf(w, "Environment: Production")
}
