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

type RecordReq struct {
	Patienid     int       `json:"patientid" validate:"required"`
	Doctorid     int       `json:"doctorid" validate:"required"`
	Date         time.Time `json:"date" validate:"required"`
	Diagnosis    string    `json:"diagnosis" validate:"required"`
	Disease      string    `json:"disease" validate:"required"`
	Prescription string    `json:"prescription" validate:"required"`
	Weight       string    `json:"weight" validate:"required"`
}

func (server *Server) createpatientrecord(w http.ResponseWriter, r *http.Request) {
	var req RecordReq
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
		server.Log.Error(err, "some error happened!")
		return
	}
	record := models.Patientrecords{
		Patienid:     req.Patienid,
		Doctorid:     req.Doctorid,
		Date:         time.Now(),
		Diagnosis:    req.Diagnosis,
		Disease:      req.Disease,
		Prescription: req.Prescription,
		Weight:       req.Weight,
	}
	record, err = server.Services.PatientRecordService.Create(record)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		server.Log.Debug(err.Error(), r.URL.Path)
		return
	}
	serializeResponse(w, http.StatusOK, record)
}

func (server *Server) updatepatientrecords(w http.ResponseWriter, r *http.Request) {
	var req RecordReq
	err := decodejson(w, r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
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
	record := models.Patientrecords{
		Recordid:     idparam,
		Diagnosis:    req.Diagnosis,
		Disease:      req.Disease,
		Prescription: req.Prescription,
		Weight:       req.Weight,
	}
	newrecord, err := server.Services.PatientRecordService.Update(record)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	serializeResponse(w, http.StatusOK, newrecord)
	log.Print("Success! ", idparam, " was updated")
}

func (server *Server) deletepatientrecord(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	err = server.Services.PatientRecordService.Delete(idparam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	serializeResponse(w, http.StatusOK, "schedule deleted successfully")
	log.Print("Success! ", idparam, " was deleted")
}
func (server *Server) findpatientrecord(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		server.Log.Error(err, r.URL.Path)
		return
	}
	record, err := server.Services.PatientRecordService.Find(idparam)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, err.Error(), http.StatusBadRequest)
			server.Log.Debug(err.Error(), r.URL.Path)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		server.Log.Debug(err.Error(), r.URL.Path)
		return
	}
	serializeResponse(w, http.StatusOK, record)
	server.Log.Info(fmt.Sprintf("record with id:%d received", record.Recordid), r.URL.Path)
}

// TODO:Error handling and logs
func (server *Server) findallpatientrecords(w http.ResponseWriter, r *http.Request) {
	page_id := r.URL.Query().Get("page_id")
	page_size := r.URL.Query().Get("page_size")
	pageid, _ := strconv.Atoi(page_id)
	if pageid < 1 {
		http.Error(w, "Page id can't be less than 1", http.StatusBadRequest)
		return
	}
	pagesize, _ := strconv.Atoi(page_size)
	skip := (pageid - 1) * pagesize
	args := models.ListPatientRecords{
		Limit:  pagesize,
		Offset: skip,
	}
	records, err := server.Services.PatientRecordService.FindAll(args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	serializeResponse(w, http.StatusOK, records)
	log.Print("Success! ", len(records), " request")
}
