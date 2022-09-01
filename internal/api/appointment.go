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
)

type AppointmentReq struct {
	Doctorid        int    `json:"doctorid" validate:"required"`
	Patientid       int    `json:"patientid" validate:"required"`
	Appointmentdate string `json:"appointmentdate" validate:"required"`
	Duration        string `json:"duration" validate:"required"`
	Approval        string `json:"approval" validate:"required"`
}

func (server *Server) createappointmentbydoctor(w http.ResponseWriter, r *http.Request) {
	var req AppointmentReq
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
	appointmentdate, err := time.Parse("2006-01-02 15:04", req.Appointmentdate)
	if err != nil {
		log.Print(err)
	}
	value, _ := checkboolfield(req.Approval)
	appointment := models.Appointment{
		Doctorid:        req.Doctorid,
		Patientid:       req.Patientid,
		Appointmentdate: appointmentdate,
		Duration:        req.Duration,
		Approval:        value,
	}
	appointment, err = server.Services.DoctorBookAppointment(appointment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		server.Log.PrintError(err, fmt.Sprintf("Agent: %s, URL: %s", r.UserAgent(), r.URL.Path), fmt.Sprintf("ResponseCode:%d", http.StatusBadRequest))
		return
	}
	fmt.Println(appointment)
	server.serializeResponse(w, http.StatusOK, appointment)
}

func (server *Server) createappointmentbypatient(w http.ResponseWriter, r *http.Request) {
	var req AppointmentReq
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
	appointmentdate, err := time.Parse("2006-01-02 15:04", req.Appointmentdate)
	if err != nil {
		log.Print(err)
	}
	appointment := models.Appointment{
		Doctorid:        req.Doctorid,
		Patientid:       req.Patientid,
		Appointmentdate: appointmentdate,
		Duration:        req.Duration,
		Approval:        false,
	}
	appointment, err = server.Services.PatientBookAppointment(appointment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		server.Log.PrintError(err, fmt.Sprintf("Agent: %s, URL: %s", r.UserAgent(), r.URL.Path), fmt.Sprintf("ResponseCode:%d", http.StatusBadRequest))
		return
	}
	server.serializeResponse(w, http.StatusOK, appointment)
}

func (server *Server) updateappointmentbyPatient(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	patient_id := params["patientid"]
	patientid, err := strconv.Atoi(patient_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	var req AppointmentReq
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
	appointmentdate, err := time.Parse("2006-01-02 15:04", req.Appointmentdate)
	if err != nil {
		log.Print(err)
	}
	value, _ := checkboolfield(req.Approval)
	appointment := models.Appointment{
		Doctorid:        req.Doctorid,
		Patientid:       patientid,
		Appointmentid:   idparam,
		Appointmentdate: appointmentdate,
		Duration:        req.Duration,
		Approval:        value,
	}
	appointment, err = server.Services.UpdateappointmentbyPatient(patientid, appointment)
	fmt.Println(appointment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	server.serializeResponse(w, http.StatusOK, appointment)
	log.Print("Success! ", appointment.Appointmentid, " was updated")
}

func (server *Server) updateappointmentbyDoctor(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	doctor_id := params["doctorid"]
	docid, err := strconv.Atoi(doctor_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	var req AppointmentReq
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
	appointmentdate, err := time.Parse("2006-01-02 15:04", req.Appointmentdate)
	if err != nil {
		log.Print(err)
	}
	value, _ := checkboolfield(req.Approval)
	appointment := models.Appointment{
		Doctorid:        req.Doctorid,
		Patientid:       req.Patientid,
		Appointmentid:   idparam,
		Appointmentdate: appointmentdate,
		Duration:        req.Duration,
		Approval:        value,
	}
	appointment, err = server.Services.UpdateappointmentbyDoctor(docid, appointment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	server.serializeResponse(w, http.StatusOK, appointment)
	log.Print("Success! ", appointment.Appointmentid, " was updated")
}

func (server *Server) deleteappointment(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	err = server.Services.AppointmentService.Delete(idparam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	server.serializeResponse(w, http.StatusOK, "schedule deleted successfully")
	log.Print("Success! ", idparam, " was deleted")
}
func (server *Server) findappointment(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	appointment, err := server.Services.AppointmentService.Find(idparam)
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
	server.serializeResponse(w, http.StatusOK, appointment)
	log.Print("Success! ", appointment.Appointmentid, " was found")
}

func (server *Server) findallappointments(w http.ResponseWriter, r *http.Request) {
	page_id := r.URL.Query().Get("page_id")
	page_size := r.URL.Query().Get("page_size")
	pageid, _ := strconv.Atoi(page_id)
	if pageid < 1 {
		http.Error(w, "Page id can't be less than 1", http.StatusBadRequest)
		return
	}
	pagesize, _ := strconv.Atoi(page_size)
	skip := (pageid - 1) * pagesize
	args := models.ListAppointments{
		Limit:  pagesize,
		Offset: skip,
	}
	appointments, err := server.Services.AppointmentService.FindAll(args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	server.serializeResponse(w, http.StatusOK, appointments)
	log.Print("Success! ", len(appointments), " request")
}
