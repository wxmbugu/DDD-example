package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/services"
	"gopkg.in/go-playground/validator.v9"
)

// TODO:Enum type for Bloodgroup i.e: A,B,AB,O
// TODO: Salt password
// TODO: Password updated at field
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
	Username   string `json:"username" validate:"required"`
	Full_name  string `json:"fullname" validate:"required"`
	Email      string `json:"email" validate:"required,email"`
	Dob        string `json:"dob" validate:"required"`
	Contact    string `json:"contact" validate:"required"`
	Bloodgroup string `json:"bloodgroup" validate:"required"`
}

//TODO: set env of tokenduration

const tokenduration = 45

func PatientResponse(patient models.Patient) PatientResp {
	return PatientResp{
		Username:   patient.Username,
		Full_name:  patient.Full_name,
		Email:      patient.Email,
		Dob:        patient.Dob.String(),
		Contact:    patient.Contact,
		Bloodgroup: patient.Bloodgroup,
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

	var req PatientLoginreq
	err := decodejson(w, r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		server.Log.Debug(err.Error(), r.URL.Path)
		return
	}
	validate := validator.New()
	err = validate.Struct(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		server.Log.Debug(err.Error(), r.URL.Path)
		return
	}

	patient, err := server.Services.PatientService.FindbyEmail(req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		server.Log.Debug(err.Error(), fmt.Sprintf("ResponseCode:%d", http.StatusBadRequest))
		return
	}
	err = services.CheckPassword(patient.Hashed_password, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		server.Log.Debug(err.Error(), r.URL.Path)
	}
	token, err := server.Auth.CreateToken(patient.Username, time.Duration(tokenduration))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		server.Log.Fatal(err, r.URL.Path)
	}
	patientres := PatientResponse(patient)
	resp := PatientLoginResp{
		AccessToken: token,
		Patient:     patientres,
	}
	serializeResponse(w, http.StatusOK, resp)
}

func (server *Server) createpatient(w http.ResponseWriter, r *http.Request) {
	var req Patientreq
	err := decodejson(w, r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		server.Log.Error(err, fmt.Sprintf("Agent: %s, URL: %s", r.UserAgent(), r.URL.Path), fmt.Sprintf("ResponseCode:%d", http.StatusBadRequest))
		return
	}
	validate := validator.New()
	err = validate.Struct(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		server.Log.Error(err, "some error happened!")
		return
	}
	dob, err := time.Parse("2006-01-02", req.Dob)
	if err != nil {
		log.Println(err)
	}
	patient := models.Patient{
		Username:        req.Username,
		Full_name:       req.Full_name,
		Email:           req.Email,
		Dob:             dob,
		Contact:         req.Contact,
		Bloodgroup:      req.Bloodgroup,
		Hashed_password: req.Hashed_password,
		Created_at:      time.Now(),
	}
	patient, err = server.Services.PatientService.Create(patient)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		server.Log.Error(err, fmt.Sprintf("Agent: %s, URL: %s", r.UserAgent(), r.URL.Path), fmt.Sprintf("ResponseCode:%d", http.StatusBadRequest))
		return
	}
	serializeResponse(w, http.StatusOK, patient)
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
