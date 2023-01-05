package api

import (
	"database/sql"
	"github.com/gorilla/mux"
	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/services"
	"gopkg.in/go-playground/validator.v9"
	"log"
	"net/http"
	"strconv"
	"time"
)

// TODO:Enum type for Bloodgroup i.e: A,B,AB,O
type Patientreq struct {
	Username        string `json:"username" validate:"required"`
	Full_name       string `json:"fullname" validate:"required"`
	Email           string `json:"email" validate:"required,email"`
	Dob             string `json:"dob" validate:"required"`
	Contact         string `json:"contact" validate:"required"`
	Bloodgroup      string `json:"bloodgroup" validate:"required"`
	Hashed_password string `json:"password" validate:"required,min=8"`
}

type PatientResp struct {
	Id        int
	Username  string `json:"username" validate:"required"`
	Full_name string `json:"fullname" validate:"required"`
}

//TODO: set env of tokenduration

const tokenduration = 45

func PatientResponse(patient models.Patient) PatientResp {
	return PatientResp{
		Username:  patient.Username,
		Full_name: patient.Full_name,
		Id:        patient.Patientid,
	}
}

type PatientLoginResp struct {
	AccessToken string      `json:"access_token"`
	Patient     PatientResp `json:"patient"`
}
type PatientLoginreq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (server *Server) PatientLogin(w http.ResponseWriter, r *http.Request) {
	var msg Form
	session, err := server.Store.Get(r, "user-session")
	if err = session.Save(r, w); err != nil {
		log.Println(err)
		http.Redirect(w, r, "/500", 300)

	}
	login := Login{
		Email:    r.PostFormValue("email"),
		Password: r.PostFormValue("password"),
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "login.html", nil)
		return
	}
	msg = Form{
		Data: &login,
	}
	if ok := msg.Validate(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "login.html", msg)
		return
	}
	patient, err := server.Services.PatientService.FindbyEmail(login.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusBadRequest)
			msg.Errors["Login"] = "No such user"
			server.Templates.Render(w, "login.html", msg)
			return
		}
		http.Redirect(w, r, "/500", 300)
	}
	if err = services.CheckPassword(patient.Hashed_password, login.Password); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Login"] = "No such user"
		server.Templates.Render(w, "login.html", msg)
		return
	}
	user := PatientResponse(patient)
	gobRegister(user)
	session.Values["user"] = user
	if err = session.Save(r, w); err != nil {
		log.Println(err)
		http.Redirect(w, r, "/500", 300)

	}
	http.Redirect(w, r, "/home", 300)
}

func (server *Server) createpatient(w http.ResponseWriter, r *http.Request) {
	var msg Form
	register := Register{
		Email:           r.PostFormValue("Email"),
		Password:        r.PostFormValue("Password"),
		ConfirmPassword: r.PostFormValue("ConfirmPassword"),
		Username:        r.PostFormValue("Username"),
		Fullname:        r.PostFormValue("Fullname"),
		Contact:         r.PostFormValue("Contact"),
		Dob:             r.PostFormValue("Dob"),
		Bloodgroup:      r.PostFormValue("Bloodgroup"),
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "register.html", nil)
		return
	}
	msg = Form{
		Data: &register,
	}
	if ok := msg.Validate(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "register.html", msg)
		return
	}
	dob, _ := time.Parse("2006-01-02", register.Dob)

	hashed_password, _ := services.HashPassword(register.Password)
	patient := models.Patient{
		Username:        register.Username,
		Full_name:       register.Fullname,
		Email:           register.Email,
		Dob:             dob,
		Contact:         register.Contact,
		Bloodgroup:      register.Bloodgroup,
		Hashed_password: hashed_password,
		Created_at:      time.Now(),
	}
	if _, err := server.Services.PatientService.Create(patient); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Exists"] = "User already Exists"
		server.Templates.Render(w, "register.html", msg)
		return
	}
	http.Redirect(w, r, "/login", 300)
}

func (server *Server) updatepatient(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	var req Patientreq
	err = decodejson(w, r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	validate := validator.New()
	err = validate.Struct(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	dob, err := time.Parse("2006-01-02", req.Dob)
	if err != nil {
		log.Println(err)
	}
	patient := models.Patient{
		Patientid:       idparam,
		Username:        req.Username,
		Full_name:       req.Full_name,
		Email:           req.Email,
		Dob:             dob,
		Contact:         req.Contact,
		Bloodgroup:      req.Bloodgroup,
		Hashed_password: req.Hashed_password,
	}
	updatedpatient, err := server.Services.PatientService.Update(patient)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	serializeResponse(w, http.StatusOK, updatedpatient)
	log.Print("Success! ", updatedpatient.Full_name, " was updated")
}

func (server *Server) deletepatient(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	err = server.Services.PatientService.Delete(idparam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	serializeResponse(w, http.StatusOK, "patient deleted successfully")
	log.Print("Success! patient with id: ", idparam, " was deleted")
}
func (server *Server) findpatient(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	patient, err := server.Services.PatientService.Find(idparam)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err.Error(), r.URL.Path, http.StatusInternalServerError)
		return
	}
	serializeResponse(w, http.StatusOK, patient)
	log.Print("Success! patient with id: ", patient.Full_name, " was received")
}

func (server *Server) findallpatients(w http.ResponseWriter, r *http.Request) {
	page_id := r.URL.Query().Get("page_id")
	page_size := r.URL.Query().Get("page_size")
	pageid, _ := strconv.Atoi(page_id)
	if pageid < 1 {
		http.Error(w, "Page id can't be less than 1", http.StatusBadRequest)
		return
	}
	pagesize, _ := strconv.Atoi(page_size)
	skip := (pageid - 1) * pagesize
	listpatients := models.ListPatients{
		Limit:  pagesize,
		Offset: skip,
	}
	patient, err := server.Services.PatientService.FindAll(listpatients)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	serializeResponse(w, http.StatusOK, patient)
	log.Print("Success! ", len(patient), " request")
}

func (server *Server) findallappointmentsbypatient(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	schedules, err := server.Services.AppointmentService.FindAllByPatient(idparam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	serializeResponse(w, http.StatusOK, schedules)

	log.Print("Success! ", len(schedules), " request")
}

func (server *Server) findallrecordsbypatient(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	records, err := server.Services.PatientRecordService.FindAllByPatient(idparam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	serializeResponse(w, http.StatusOK, records)
	log.Print("Success! ", len(records), " request")
}
