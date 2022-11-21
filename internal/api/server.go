package api

import (
	"fmt"
	"net/http"

	//:w"strings"

	"github.com/gorilla/mux"
	"github.com/patienttracker/internal/auth"
	"github.com/patienttracker/internal/services"
	"github.com/patienttracker/pkg/logger"
)

// TODO: admin & admin Templates.

const version = "1.0.0"

type Server struct {
	Router   *mux.Router
	Services services.Service
	Log      *logger.Logger
	Auth     auth.Token
}

func NewServer(services services.Service, router *mux.Router) *Server {

	logger := logger.New()
	token, err := auth.PasetoMaker("YELLOW SUBMARINE, BLACK WIZARDRY")
	if err != nil {
		logger.Debug(err.Error())
	}
	server := Server{
		Router:   router,
		Log:      logger,
		Services: services,
		Auth:     token,
	}
	server.Routes()

	return &server
}

func (server *Server) Routes() {
	/* http.Handle("/", server.Router) */
	server.Router.Use(jsonmiddleware)
	//server.Router.Use(server.contentTypeMiddleware)
	server.Router.HandleFunc("/v1/healthcheck", server.Healthcheck).Methods("GET")
	server.Router.HandleFunc("/v1/department", server.createdepartment).Methods("POST")
	server.Router.HandleFunc("/v1/department/{id:[0-9]+}", server.finddepartment).Methods("GET")
	server.Router.HandleFunc("/v1/department/delete/{id:[0-9]+}", server.deletedepartment).Methods("DELETE")
	//queryparams: ->page_id && page_size
	server.Router.HandleFunc("/v1/departments", server.findalldepartment).Methods("GET")
	server.Router.HandleFunc("/v1/department/update/{id:[0-9]+}", server.updatedepartment).Methods("POST")
	//queryparams: ->page_id && page_size
	server.Router.HandleFunc("/v1/{departmentname}", server.findalldoctorsbydepartment).Methods("GET")

	server.Router.HandleFunc("/v1/doctor", server.createdoctor).Methods("POST")
	server.Router.HandleFunc("/v1/patient", server.createpatient).Methods("POST")
	server.Router.HandleFunc("/v1/patient/login", server.PatientLogin).Methods("POST")
	server.Router.HandleFunc("/v1/doctor/login", server.DoctorLogin).Methods("POST")

	// auth middleware
	authroutes := server.Router.PathPrefix("/v1").Subrouter()
	authroutes.Use(server.authmiddleware)

	authroutes.HandleFunc("/doctor/{id:[0-9]+}", server.finddoctor).Methods("GET")
	authroutes.HandleFunc("/doctor/{id:[0-9]+}", server.deletedoctor).Methods("DELETE")
	//queryparams: ->page_i && page_size
	authroutes.HandleFunc("/doctors", server.findalldoctors).Methods("GET")
	authroutes.HandleFunc("/doctor/{id:[0-9]+}", server.updatedoctor).Methods("POST")
	authroutes.HandleFunc("/doctor/{id:[0-9]+}/schedules", server.findallschedulesbydoctor).Methods("GET")
	authroutes.HandleFunc("/doctor/{id:[0-9]+}/appointments", server.findallappointmentsbydoctor).Methods("GET")
	authroutes.HandleFunc("/doctor/{id:[0-9]+}/records", server.findallrecordsbydoctor).Methods("GET")

	authroutes.HandleFunc("/patient", server.findpatient).Methods("GET")

	authroutes.HandleFunc("/patient/{id:[0-9]+}", server.deletepatient).Methods("DELETE")
	authroutes.HandleFunc("/patients", server.findallpatients).Methods("GET")
	authroutes.HandleFunc("/patient/{id:[0-9]+}", server.updatepatient).Methods("POST")
	authroutes.HandleFunc("/patient/{id:[0-9]+}/appoinmtents", server.findallappointmentsbypatient).Methods("GET")
	authroutes.HandleFunc("/patient/{id:[0-9]+}/records", server.findallrecordsbypatient).Methods("GET")

	authroutes.HandleFunc("/schedule", server.createschedule).Methods("POST")
	authroutes.HandleFunc("/schedule", server.findschedule).Methods("GET")
	authroutes.HandleFunc("/schedule/{id:[0-9]+}", server.deleteschedule).Methods("DELETE")
	authroutes.HandleFunc("/schedules", server.findallschedules).Methods("GET")
	authroutes.HandleFunc("/schedule/{id:[0-9]+}", server.updateschedule).Methods("POST")

	authroutes.HandleFunc("/appointment/patient/{id:[0-9]+}", server.createappointmentbypatient).Methods("POST")
	authroutes.HandleFunc("/appointment/doctor/{id:[0-9]+}", server.createappointmentbydoctor).Methods("POST")
	authroutes.HandleFunc("/appointment/{id:[0-9]+}", server.findappointment).Methods("GET")
	authroutes.HandleFunc("/appointment/{id:[0-9]+}", server.deleteappointment).Methods("DELETE")
	authroutes.HandleFunc("/appointments", server.findallappointments).Methods("GET")
	authroutes.HandleFunc("/appointment/doctor", server.UpdateDoctorAppointment).Methods("POST")
	authroutes.HandleFunc("/appointment/patient", server.updateappointmentbyPatient).Methods("POST")

	authroutes.HandleFunc("/record", server.createpatientrecord).Methods("POST")
	authroutes.HandleFunc("/record", server.findpatientrecord).Methods("GET")
	authroutes.HandleFunc("/record/{id:[0-9]+}", server.deletepatientrecord).Methods("DELETE")
	authroutes.HandleFunc("/records", server.findallpatientrecords).Methods("GET")
	authroutes.HandleFunc("/record/{id:[0-9]+}", server.updatepatientrecords).Methods("POST")
}
func (server *Server) Healthcheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "status: available\n")
	fmt.Fprintf(w, "version: %s\n", version)
	fmt.Fprintf(w, "Environment: Production")
}
