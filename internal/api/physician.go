package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/services"
)

type Doctorreq struct {
	Username        string `json:"username" validate:"required"`
	Full_name       string `json:"fullname" validate:"required"`
	Email           string `json:"email" validate:"required,email"`
	Contact         string `json:"contact" validate:"required"`
	Hashed_password string `json:"password" validate:"required,min=8"`
	Departmentname  string `json:"departmentname" validate:"required"`
}

func (server *Server) createdoctor(w http.ResponseWriter, r *http.Request) {
	var req Doctorreq
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
	doctor := models.Physician{
		Username:        req.Username,
		Full_name:       req.Full_name,
		Email:           req.Email,
		Contact:         req.Contact,
		Hashed_password: req.Hashed_password,
		Created_at:      time.Now(),
		Departmentname:  req.Departmentname,
	}
	doctor, err = server.Services.DoctorService.Create(doctor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		server.Log.Error(err, fmt.Sprintf("Agent: %s, URL: %s", r.UserAgent(), r.URL.Path), fmt.Sprintf("ResponseCode:%d", http.StatusBadRequest))
		return
	}
	serializeResponse(w, http.StatusOK, doctor)
}

type DoctorResp struct {
	Username       string `json:"username"`
	Full_name      string `json:"fullname"`
	Email          string `json:"email"`
	Contact        string `json:"contact"`
	Departmentname string `json:"departmentname"`
}

func DoctorResponse(doctor models.Physician) DoctorResp {
	return DoctorResp{
		Username:       doctor.Username,
		Full_name:      doctor.Full_name,
		Email:          doctor.Email,
		Contact:        doctor.Contact,
		Departmentname: doctor.Departmentname,
	}
}

type DoctorLoginResp struct {
	AccessToken string     `json:"access_token"`
	Doctor      DoctorResp `json:"doctor"`
}
type DoctorLoginreq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (server *Server) DoctorLogin(w http.ResponseWriter, r *http.Request) {
	var req DoctorLoginreq
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

	doctor, err := server.Services.DoctorService.FindbyEmail(req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		server.Log.Debug(err.Error(), fmt.Sprintf("ResponseCode:%d", http.StatusBadRequest))
		return
	}
	err = services.CheckPassword(doctor.Hashed_password, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		server.Log.Debug(err.Error(), r.URL.Path)
	}
	token, err := server.Auth.CreateToken(doctor.Username, time.Duration(tokenduration))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		server.Log.Fatal(err, r.URL.Path)
	}
	docresp := DoctorResponse(doctor)
	resp := DoctorLoginResp{
		AccessToken: token,
		Doctor:      docresp,
	}
	serializeResponse(w, http.StatusOK, resp)
}

func (server *Server) updatedoctor(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	var req Doctorreq
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
	doctor := models.Physician{
		Physicianid:     idparam,
		Username:        req.Username,
		Full_name:       req.Full_name,
		Email:           req.Email,
		Contact:         req.Contact,
		Hashed_password: req.Hashed_password,
		Departmentname:  req.Departmentname,
	}
	updateddoctor, err := server.Services.DoctorService.Update(doctor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	serializeResponse(w, http.StatusOK, updateddoctor)
	log.Print("Success! ", updateddoctor, " was updated")
}

func (server *Server) deletedoctor(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	err = server.Services.DoctorService.Delete(idparam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	serializeResponse(w, http.StatusOK, "doctor deleted successfully")
	log.Print("Success! doctor with id: ", idparam, " was deleted")
}

func (server *Server) finddoctor(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	doc, err := server.Services.DoctorService.Find(idparam)
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
	serializeResponse(w, http.StatusOK, doc)
	log.Print("Success! doctor with id: ", doc.Username, " was received")
}

func (server *Server) findalldoctors(w http.ResponseWriter, r *http.Request) {
	page_id := r.URL.Query().Get("page_id")
	page_size := r.URL.Query().Get("page_size")
	pageid, _ := strconv.Atoi(page_id)
	if pageid < 1 {
		http.Error(w, "Page id can't be less than 1", http.StatusBadRequest)
		return
	}
	pagesize, _ := strconv.Atoi(page_size)
	skip := (pageid - 1) * pagesize
	listdoctors := models.ListDoctors{
		Limit:  pagesize,
		Offset: skip,
	}
	departments, err := server.Services.DoctorService.FindAll(listdoctors)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	serializeResponse(w, http.StatusOK, departments)
	log.Print("Success! ", len(departments), " request")
}

func (server *Server) findallschedulesbydoctor(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	schedules, err := server.Services.ScheduleService.FindbyDoctor(idparam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	serializeResponse(w, http.StatusOK, schedules)
	log.Print("Success! ", len(schedules), " request")
}

func (server *Server) findallappointmentsbydoctor(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	schedules, err := server.Services.AppointmentService.FindAllByDoctor(idparam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	serializeResponse(w, http.StatusOK, schedules)
	log.Print("Success! ", len(schedules), " request")
}

func (server *Server) findallrecordsbydoctor(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	records, err := server.Services.PatientRecordService.FindAllByDoctor(idparam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	serializeResponse(w, http.StatusOK, records)
	log.Print("Success! ", len(records), " request")
}
