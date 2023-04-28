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

// TODO: admin Templates. - Users
// PERF: <WIP:Needs to be tested out> Create appoinmtent & Book appointment will be slow when a user has many appointments <module:Services>
// TODO: Delete <modal are you sure?????>
// TODO: Search functionality
// TODO: Add Textbox for children under 18 years && guardian details
// TODO: Add Nurse
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
	redis := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
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
	server.Router.HandleFunc("/500", server.InternalServeError)
	server.Router.HandleFunc("/v1/healthcheck", server.Healthcheck).Methods("GET")
	server.Router.HandleFunc("/verify/{id}", server.VerifyAccount)
	server.Router.HandleFunc("/register", server.createpatient)
	server.Router.HandleFunc("/login", server.PatientLogin)
	server.Router.HandleFunc("/admin/login", server.AdminLogin)
	server.Router.HandleFunc("/staff/login", server.StaffLogin)
	// staff i.e Doctors
	staff := server.Router.PathPrefix("/staff").Subrouter()
	staff.Use(server.sessionstaffmiddleware)
	staff.HandleFunc("/home", server.Staffhome)
	staff.HandleFunc("/logout", server.StaffLogout)
	staff.HandleFunc("/records", server.Staffrecord)
	staff.HandleFunc("/appointments", server.Staffappointments)
	staff.HandleFunc("/schedules", server.Staffschedule)
	staff.HandleFunc("/update/appointment/{id:[0-9]+}", server.StaffUpdateAppointment)
	staff.HandleFunc("/register/record/{id:[0-9]+}", server.StaffCreateRecord)
	staff.HandleFunc("/register/schedule", server.Staffcreateschedule)
	staff.HandleFunc("/update/schedule/{id:[0-9]+}", server.Staffupdateschedule)
	staff.HandleFunc("/update/record/{id:[0-9]+}", server.StaffUpdateRecord)
	staff.HandleFunc("/delete/schedule/{id:[0-9]+}", server.Staffdeleteschedule)
	staff.HandleFunc("/profile", server.Staffprofile)

	admin := server.Router.PathPrefix("/admin").Subrouter()
	admin.Use(server.sessionadminmiddleware)
	admin.HandleFunc("/home", server.Adminhome)
	admin.HandleFunc("/logout", server.AdminLogout)
	admin.HandleFunc("/records/{pageid:[0-9]+}", server.Adminrecord)
	admin.HandleFunc("/appointments/{pageid:[0-9]+}", server.Adminappointments)
	admin.HandleFunc("/users", server.Adminuser)
	admin.HandleFunc("/roles", server.Adminroles)
	admin.HandleFunc("/patients/{pageid:[0-9]+}", server.Adminpatient)
	admin.HandleFunc("/physician/{pageid:[0-9]+}", server.Adminphysician)
	admin.HandleFunc("/schedule/{pageid:[0-9]+}", server.Adminschedule)
	admin.HandleFunc("/department/{pageid:[0-9]+}", server.Admindepartment)
	admin.HandleFunc("/register/patient", server.Admincreatepatient)
	admin.HandleFunc("/register/user", server.Admincreateuser)
	admin.HandleFunc("/register/doctor", server.Admincreatedoctor)
	admin.HandleFunc("/register/department", server.Admincreatedepartment)
	admin.HandleFunc("/register/record", server.Admincreaterecords)
	admin.HandleFunc("/register/appointment", server.AdmincreateAppointment)
	admin.HandleFunc("/register/schedule", server.Admincreateschedule)
	admin.HandleFunc("/register/roles", server.AdmincreateRoles)
	admin.HandleFunc("/delete/patient/{id:[0-9]+}", server.Admindeletepatient)
	admin.HandleFunc("/delete/doctor/{id:[0-9]+}", server.Admindeletedoctor)
	admin.HandleFunc("/delete/user/{id:[0-9]+}", server.Admindeleteuser)
	admin.HandleFunc("/delete/role/{id:[0-9]+}", server.Admindeleterole)
	admin.HandleFunc("/delete/department/{id:[0-9]+}", server.Admindeletedepartment)
	admin.HandleFunc("/delete/record/{id:[0-9]+}", server.Admindeleterecord)
	admin.HandleFunc("/delete/appointment/{id:[0-9]+}", server.Admindeleteappointment)
	admin.HandleFunc("/delete/schedule/{id:[0-9]+}", server.Admindeleteschedule)
	admin.HandleFunc("/update/patient/{id:[0-9]+}", server.Adminupdatepatient)
	admin.HandleFunc("/update/user/{id:[0-9]+}", server.Adminupdateuser)
	admin.HandleFunc("/update/record/{id:[0-9]+}", server.Adminupdaterecords)
	admin.HandleFunc("/update/role/{id:[0-9]+}", server.Adminupdateroles)
	admin.HandleFunc("/update/doctor/{id:[0-9]+}", server.Adminupdatedoctor)
	admin.HandleFunc("/update/doctor/{id:[0-9]+}", server.Adminupdatedoctor)
	admin.HandleFunc("/update/appointment/{id:[0-9]+}", server.AdminupdateAppointment)
	admin.HandleFunc("/update/schedule/{id:[0-9]+}", server.Adminupdateschedule)
	admin.HandleFunc("/update/department/{id:[0-9]+}", server.Adminupdatedepartment)

	// session middleware
	session := server.Router.PathPrefix("/").Subrouter()
	session.Use(server.sessionmiddleware)
	session.HandleFunc("/home", server.home)
	session.HandleFunc("/logout", server.PatientLogout)
	session.HandleFunc("/records", server.record)
	session.HandleFunc("/appointments", server.appointments)
	session.HandleFunc("/departments", server.Patientshowdepartments)
	session.HandleFunc("/department/{name}/doctors", server.PatientListDoctorsDept)
	session.HandleFunc("/appointment/doctor/{id:[0-9]+}", server.PatienBookAppointment)
	session.HandleFunc("/update/appointment/{id:[0-9]+}", server.PatienUpdateAppointment)
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
	return
}
func (server *Server) Homepage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "startpage.html", nil)
	return
}
func gobRegister(data any) {
	gob.Register(data)
}
