package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	"github.com/patienttracker/internal/models"
)

// TODO:Enum type for Bloodgroup i.e: A,B,AB,O
// TODO: Salt password
// TODO: Password updated at field
type Patientreq struct {
	Username            string    `json:"username" validate:"required"`
	Full_name           string    `json:"fullname" validate:"required"`
	Email               string    `json:"email" validate:"required,email"`
	Dob                 string    `json:"dob" validate:"required"`
	Contact             string    `json:"contact" validate:"required"`
	Bloodgroup          string    `json:"bloodgroup" validate:"required"`
	Hashed_password     string    `json:"password" validate:"required"`
	Password_changed_at time.Time `json:"password_changed_at" validate:"required"`
	Created_at          time.Time `json:"created_at" validate:"required"`
}

func (server *Server) createpatient(w http.ResponseWriter, r *http.Request) {
	var req Patientreq
	err := decodejson(w, r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		server.Log.PrintError(err, fmt.Sprintf("Agent: %s, URL: %s", r.UserAgent(), r.URL.Path), fmt.Sprintf("ResponseCode:%d", http.StatusBadRequest))
		return
	}
	validate := validator.New()
	err = validate.Struct(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		server.Log.PrintError(err, "some error happened!")
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
		server.Log.PrintError(err, fmt.Sprintf("Agent: %s, URL: %s", r.UserAgent(), r.URL.Path), fmt.Sprintf("ResponseCode:%d", http.StatusBadRequest))
		return
	}
	server.serializeResponse(w, http.StatusOK, patient)
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
	server.serializeResponse(w, http.StatusOK, updatedpatient)
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
	server.serializeResponse(w, http.StatusOK, "patient deleted successfully")
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	server.serializeResponse(w, http.StatusOK, patient)
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
	server.serializeResponse(w, http.StatusOK, patient)
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
	server.serializeResponse(w, http.StatusOK, schedules)
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
	server.serializeResponse(w, http.StatusOK, records)
	log.Print("Success! ", len(records), " request")
}
