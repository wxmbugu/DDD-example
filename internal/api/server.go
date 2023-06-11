package api

import (
	"context"
	"encoding/gob"
	"fmt"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/patienttracker/internal/auth"
	"github.com/patienttracker/internal/services"
	"github.com/patienttracker/internal/worker"
	"github.com/patienttracker/pkg/logger"
	tmp "github.com/patienttracker/template"
	"github.com/redis/go-redis/v9"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync"
)

// TODO: Calendar
// TODO: Documentation
// TODO: Slides
const version = "1.0.0"

type Server struct {
	Router    *mux.Router
	Services  *services.Service
	Log       *logger.Logger
	Auth      auth.Token
	Templates tmp.Template
	Store     *sessions.CookieStore
	Mailer    *SendEmails
	Redis     *redis.Client
	Worker    worker.Worker
	Context   context.Context
	sync.WaitGroup
}

func NewServer(services services.Service, router *mux.Router) *Server {
	logger := logger.New()
	token, err := auth.PasetoMaker("YELLOW SUBMARINE, BLACK WIZARDRY") // TODO: keep this value in env file
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
	redis := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // TODO: keep this value in env file
		Password: "",               // no password set
		DB:       0,                // use default DB
	})
	mailworker := NewSenderMail()
	workerchan := make(chan chan worker.Task, 100)
	woker := worker.Newworker(10, workerchan)
	server := Server{
		Router:    router,
		Log:       logger,
		Services:  &services,
		Auth:      token,
		Templates: *temp,
		Store:     store,
		Redis:     redis,
		Mailer:    &mailworker,
		Worker:    woker,
		Context:   context.Background(),
	}
	server.Routes()
	return &server
}

func (server *Server) Routes() {
	getwd, _ := os.Getwd()
	path := getwd + "/upload"
	fs := http.FileServer(http.FS(tmp.Static()))
	upload := http.FileServer(http.Dir(path))
	server.Router.PathPrefix("/upload/").Handler(http.StripPrefix("/upload/", upload))
	server.Router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	server.Router.Use(server.LoggingMiddleware)
	server.Router.Use(csrf.Protect([]byte("MgONCCTehPKsRZyfBsBdjdL83X7ABRkt"), csrf.SameSite(csrf.SameSiteStrictMode))) // TODO: keep this value in env file
	server.Router.HandleFunc("/", server.Homepage)
	server.Router.HandleFunc("/v1/healthcheck", server.Healthcheck).Methods("GET")
	server.Router.HandleFunc("/verify/{id}", server.VerifyAccount)
	server.Router.HandleFunc("/register", server.createpatient)
	server.Router.HandleFunc("/login", server.PatientLogin)
	server.Router.HandleFunc("/admin/login", server.AdminLogin)
	server.Router.HandleFunc("/staff/login", server.StaffLogin)
	server.Router.HandleFunc("/nurse/login", server.NurseLogin)
	server.Router.HandleFunc("/patient/forgotpassword", server.resetpassword)
	server.Router.HandleFunc("/nurse/forgotpassword", server.resetpassword)
	server.Router.HandleFunc("/admin/forgotpassword", server.resetpassword)
	server.Router.HandleFunc("/doctor/forgotpassword", server.resetpassword)
	server.Router.HandleFunc("/patient/passwordreset", server.patient_reset_password)
	server.Router.HandleFunc("/doctor/passwordreset", server.doctor_reset_password)
	server.Router.HandleFunc("/nurse/passwordreset", server.nurse_reset_password)
	server.Router.HandleFunc("/admin/passwordreset", server.admin_reset_password)
	server.Router.HandleFunc("/500", server.InternalServeError)
	server.Router.HandleFunc("/404", server.NotFound)

	staff := server.Router.PathPrefix("/staff").Subrouter()
	staff.Use(server.sessionstaffmiddleware)
	staff.HandleFunc("/home", server.Staffhome)
	staff.HandleFunc("/logout", server.StaffLogout)
	staff.HandleFunc("/records", server.Staffrecord)
	staff.HandleFunc("/appointments", server.Staffappointments)
	staff.HandleFunc("/schedules", server.Staffschedule)
	staff.HandleFunc("/update/appointment/{id:[0-9]+}", server.StaffUpdateAppointment)
	staff.HandleFunc("/view/record/{id:[0-9]+}", server.staffviewrecord)
	staff.HandleFunc("/register/schedule", server.Staffcreateschedule)
	staff.HandleFunc("/update/schedule/{id:[0-9]+}", server.Staffupdateschedule)
	staff.HandleFunc("/delete/schedule/{id:[0-9]+}", server.Staffdeleteschedule)
	staff.HandleFunc("/profile", server.Staffprofile)

	admin := server.Router.PathPrefix("/admin").Subrouter()
	admin.Use(server.sessionadminmiddleware)
	admin.HandleFunc("/home", server.Adminhome)
	admin.HandleFunc("/logout", server.AdminLogout)
	admin.HandleFunc("/records/{pageid:[0-9]+}", server.CheckPermissions(server.Adminrecord, services.Or{Permissions: []string{"admin", "viewer", "editor", "record:admin", "record:editor", "record:viewer"}}))
	admin.HandleFunc("/appointments/{pageid:[0-9]+}", server.CheckPermissions(server.Adminappointments, services.Or{Permissions: []string{"admin", "viewer", "editor", "appointment:admin", "appointment:editor", "appointment:viewer"}}))
	admin.HandleFunc("/users", server.CheckPermissions(server.Adminuser, services.Or{Permissions: []string{"admin"}}))
	admin.HandleFunc("/roles", server.CheckPermissions(server.Adminroles, services.Or{Permissions: []string{"admin"}}))
	admin.HandleFunc("/doctor", server.CheckPermissions(server.Adminfilterphysician, services.Or{Permissions: []string{"admin", "viewer", "editor", "physician:admin", "physician:editor", "physician:viewer"}}))
	admin.HandleFunc("/patient", server.CheckPermissions(server.Adminfilterpatient, services.Or{Permissions: []string{"admin", "viewer", "editor", "patient:admin", "patient:editor", "patient:viewer"}}))
	admin.HandleFunc("/nurses", server.CheckPermissions(server.Adminfilternurse, services.Or{Permissions: []string{"admin", "viewer", "editor", "nurse:admin", "nurse:editor", "nurse:viewer"}}))
	admin.HandleFunc("/schedule/{pageid:[0-9]+}", server.CheckPermissions(server.Adminschedule, services.Or{Permissions: []string{"admin", "viewer", "editor", "schedule:admin", "schedule:editor", "schedule:viewer"}}))
	admin.HandleFunc("/department/{pageid:[0-9]+}", server.CheckPermissions(server.Admindepartment, services.Or{Permissions: []string{"admin", "editor", "department:admin", "department:viewer", "department:editor"}}))
	admin.HandleFunc("/register/patient", server.CheckPermissions(server.Admincreatepatient, services.Or{Permissions: []string{"admin", "editor", "patient:admin", "patient:editor"}}))
	admin.HandleFunc("/register/user", server.CheckPermissions(server.Admincreateuser, services.Or{Permissions: []string{"admin"}}))
	admin.HandleFunc("/register/nurse", server.CheckPermissions(server.Admincreatenurse, services.Or{Permissions: []string{"admin", "editor", "nurse:admin", "nurse:editor"}}))
	admin.HandleFunc("/register/doctor", server.CheckPermissions(server.Admincreatedoctor, services.Or{Permissions: []string{"admin", "editor", "physician:admin", "physician:editor"}}))
	admin.HandleFunc("/register/department", server.CheckPermissions(server.Admincreatedepartment, services.Or{Permissions: []string{"admin", "editor", "department:admin", "department:editor"}}))
	admin.HandleFunc("/register/record", server.CheckPermissions(server.Admincreaterecords, services.Or{Permissions: []string{"admin", "editor", "record:admin", "record:editor"}}))
	admin.HandleFunc("/register/appointment", server.CheckPermissions(server.AdmincreateAppointment, services.Or{Permissions: []string{"admin", "editor", "appointment:admin", "appointment:editor"}}))
	admin.HandleFunc("/register/schedule", server.CheckPermissions(server.Admincreateschedule, services.Or{Permissions: []string{"admin", "editor", "schedule:admin", "schedule:editor"}}))
	admin.HandleFunc("/register/roles", server.CheckPermissions(server.AdmincreateRoles, services.Or{Permissions: []string{"admin"}}))
	admin.HandleFunc("/delete/patient/{id:[0-9]+}", server.CheckPermissions(server.Admindeletepatient, services.Or{Permissions: []string{"admin", "editor", "patient:admin", "patient:editor"}}))
	admin.HandleFunc("/delete/doctor/{id:[0-9]+}", server.CheckPermissions(server.Admindeletedoctor, services.Or{Permissions: []string{"admin", "editor", "physician:admin", "physician:editor"}}))
	admin.HandleFunc("/delete/user/{id:[0-9]+}", server.CheckPermissions(server.Admindeleteuser, services.Or{Permissions: []string{"admin"}}))
	admin.HandleFunc("/delete/role/{id:[0-9]+}", server.CheckPermissions(server.Admindeleterole, services.Or{Permissions: []string{"admin"}}))
	admin.HandleFunc("/delete/nurse/{id:[0-9]+}", server.CheckPermissions(server.Admindeletenurse, services.Or{Permissions: []string{"admin", "editor", "nurse:admin", "nurse:editor"}}))
	admin.HandleFunc("/delete/department/{id:[0-9]+}", server.CheckPermissions(server.Admindeletedepartment, services.Or{Permissions: []string{"admin", "editor", "department:admin", "department:editor"}}))
	admin.HandleFunc("/delete/record/{id:[0-9]+}", server.CheckPermissions(server.Admindeleterecord, services.Or{Permissions: []string{"admin", "editor", "record:admin", "record:editor"}}))
	admin.HandleFunc("/delete/appointment/{id:[0-9]+}", server.CheckPermissions(server.Admindeleteappointment, services.Or{Permissions: []string{"admin", "editor", "appointment:admin", "appointment:editor"}}))
	admin.HandleFunc("/delete/schedule/{id:[0-9]+}", server.CheckPermissions(server.Admindeleteschedule, services.Or{Permissions: []string{"admin", "editor", "schedule:admin", "schedule:editor"}}))
	admin.HandleFunc("/update/patient/{id:[0-9]+}", server.CheckPermissions(server.Adminupdatepatient, services.Or{Permissions: []string{"admin", "editor", "patient:admin", "patient:editor"}}))
	admin.HandleFunc("/update/user/{id:[0-9]+}", server.CheckPermissions(server.Adminupdateuser, services.Or{Permissions: []string{"admin"}}))
	admin.HandleFunc("/update/record/{id:[0-9]+}", server.CheckPermissions(server.Adminupdaterecords, services.Or{Permissions: []string{"admin", "editor", "record:admin", "record:editor"}}))
	admin.HandleFunc("/update/role/{id:[0-9]+}", server.CheckPermissions(server.Adminupdateroles, services.Or{Permissions: []string{"admin"}}))
	admin.HandleFunc("/update/doctor/{id:[0-9]+}", server.CheckPermissions(server.Adminupdatedoctor, services.Or{Permissions: []string{"admin", "editor", "physician:admin", "physician:editor"}}))
	admin.HandleFunc("/update/appointment/{id:[0-9]+}", server.CheckPermissions(server.AdminupdateAppointment, services.Or{Permissions: []string{"admin", "editor", "appointment:admin", "appointment:editor"}}))
	admin.HandleFunc("/update/schedule/{id:[0-9]+}", server.CheckPermissions(server.Adminupdateschedule, services.Or{Permissions: []string{"admin", "editor", "schedule:admin", "schedule:editor"}}))
	admin.HandleFunc("/update/department/{id:[0-9]+}", server.CheckPermissions(server.Adminupdatedepartment, services.Or{Permissions: []string{"admin", "editor", "department:admin", "department:editor"}}))
	admin.HandleFunc("/update/nurse/{id:[0-9]+}", server.CheckPermissions(server.Adminupdatenurse, services.Or{Permissions: []string{"admin", "editor", "nurse:admin", "nurse:editor"}}))
	nurse := server.Router.PathPrefix("/nurse").Subrouter()
	nurse.Use(server.sessionnursemiddleware)
	nurse.HandleFunc("/logout", server.NurseLogout)
	nurse.HandleFunc("/home", server.NurseCreateRecord)
	nurse.HandleFunc("/records", server.Nurserecord)
	nurse.HandleFunc("/view/record/{id:[0-9]+}", server.NurseViewRecord)

	// session middleware
	session := server.Router.PathPrefix("/").Subrouter()
	session.Use(server.sessionmiddleware)
	session.HandleFunc("/home", server.home)
	session.HandleFunc("/logout", server.PatientLogout)
	session.HandleFunc("/records", server.record)
	session.HandleFunc("/appointments", server.appointments)
	session.HandleFunc("/doctor", server.Patientfilterphysician)
	session.HandleFunc("/appointment/doctor/{id:[0-9]+}", server.PatienBookAppointment)
	session.HandleFunc("/update/appointment/{id:[0-9]+}", server.PatientUpdateAppointment)
	session.HandleFunc("/view/record/{id:[0-9]+}", server.PatientViewRecord)
	session.HandleFunc("/profile", server.profile)
}
func (server *Server) Healthcheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "status: available\n")
	fmt.Fprintf(w, "version: %s\n", version)
	fmt.Fprintf(w, "Environment: Production")
}
func (server *Server) InternalServeError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	server.Templates.Render(w, "500.html", nil)
}
func (server *Server) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	server.Templates.Render(w, "404.html", nil)
}
func (server *Server) Homepage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "startpage.html", nil)
}
func gobRegister(data any) {
	gob.Register(data)
}
