package api

import (
	"encoding/gob"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/patienttracker/internal/auth"
	"github.com/patienttracker/internal/services"
	"github.com/patienttracker/pkg/logger"
	tmp "github.com/patienttracker/template"
	"net/http"
)

// TODO: admin & admin Templates.

const version = "1.0.0"

type Server struct {
	Router    *mux.Router
	Services  services.Service
	Log       *logger.Logger
	Auth      auth.Token
	Templates tmp.Template
	Store     *sessions.CookieStore
}

func NewServer(services services.Service, router *mux.Router) *Server {
	logger := logger.New()
	token, err := auth.PasetoMaker("YELLOW SUBMARINE, BLACK WIZARDRY")
	if err != nil {
		logger.Debug(err.Error())
	}
	temp := tmp.New()
	authKey := securecookie.GenerateRandomKey(64)
	encryptionKey := securecookie.GenerateRandomKey(32)
	store := sessions.NewCookieStore(
		authKey,
		encryptionKey,
	)
	server := Server{
		Router:    router,
		Log:       logger,
		Services:  services,
		Auth:      token,
		Templates: *temp,
		Store:     store,
	}
	server.Routes()

	return &server
}

func (server *Server) Routes() {
	// contentStatic, _ := fs.Sub(static, "./static/")
	// server.Router.Handle("/", http.FileServer(http.FS(contentStatic)))
	fs := http.FileServer(http.FS(tmp.Static()))
	server.Router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	// server.Router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(tmp.Content))))
	/* http.Handle("/", server.Router) */
	// server.Router.Use(jsonmiddleware)
	//server.Router.Use(server.contentTypeMiddleware)
	server.Router.Use(server.LoggingMiddleware)

	server.Router.HandleFunc("/500", server.InternalServeError)
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
	server.Router.HandleFunc("/register", server.createpatient)
	server.Router.HandleFunc("/login", server.PatientLogin)
	server.Router.HandleFunc("/admin/login", server.AdminLogin)

	// session middleware
	session := server.Router.PathPrefix("/").Subrouter()
	session.Use(server.sessionmiddleware)
	session.HandleFunc("/home", server.home)
	session.HandleFunc("/records", server.record)
	session.HandleFunc("/appointments", server.appointments)
	// staff session
	server.Router.HandleFunc("/staff/login", server.DoctorLogin).Methods("POST")
	//auth
	admin := server.Router.PathPrefix("/admin").Subrouter()
	admin.Use(server.sessionmiddleware)
	admin.HandleFunc("/home", server.Adminhome)
	// admin.HandleFunc("/user", server.Adminuser)
	// admin.HandleFunc("/roles", server.Adminroles)
	// admin.HandleFunc("/permissions", server.Adminpermissions)
	// admin.HandleFunc("/patient", server.Adminpatient)
	// admin.HandleFunc("/physician", server.Adminphysician)
	// admin.HandleFunc("/schedule", server.Adminschedule)
	// admin.HandleFunc("/appointment", server.Adminappointment)
	// admin.HandleFunc("/records", server.Adminrecords)
	// admin.HandleFunc("/department", server.Admindepartment)

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
func (server *Server) InternalServeError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	server.Templates.Render(w, "500.html", nil)
	return
}

func gobRegister(data any) {
	gob.Register(data)
}
