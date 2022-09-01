package api

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	"github.com/patienttracker/internal/models"
)

type ScheduleReq struct {
	Doctorid  int    `json:"doctorid" validate:"required"`
	Starttime string `json:"starttime" validate:"required"`
	Endtime   string `json:"endtime" validate:"required"`
	Active    string `json:"active" validate:"required"`
}

func checkboolfield(data any) (bool, error) {
	if data == "true" || data == 1 {
		return true, nil
	} else if data == "false" || data == 0 {
		return false, nil
	}
	return false, errors.New("unkown field type")
}

func (server *Server) createschedule(w http.ResponseWriter, r *http.Request) {
	var req ScheduleReq
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
	active, _ := checkboolfield(req.Active)
	schedule := models.Schedule{
		Doctorid:  req.Doctorid,
		Starttime: req.Starttime,
		Endtime:   req.Endtime,
		Active:    active,
	}
	schedule, err = server.Services.MakeSchedule(schedule)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		server.Log.PrintError(err, fmt.Sprintf("Agent: %s, URL: %s", r.UserAgent(), r.URL.Path), fmt.Sprintf("ResponseCode:%d", http.StatusBadRequest))
		return
	}
	server.serializeResponse(w, http.StatusOK, schedule)
}

func (server *Server) updateschedule(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	var req ScheduleReq
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
	active, _ := checkboolfield(req.Active)
	schedule := models.Schedule{
		Scheduleid: idparam,
		Doctorid:   req.Doctorid,
		Starttime:  req.Starttime,
		Endtime:    req.Endtime,
		Active:     active,
	}
	schedule, err = server.Services.UpdateSchedule(schedule)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	server.serializeResponse(w, http.StatusOK, schedule)
	log.Print("Success! ", schedule.Scheduleid, " was updated")
}

func (server *Server) deleteschedule(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	err = server.Services.ScheduleService.Delete(idparam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	server.serializeResponse(w, http.StatusOK, "schedule deleted successfully")
	log.Print("Success! ", idparam, " was deleted")
}

func (server *Server) findschedule(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	schedule, err := server.Services.ScheduleService.Find(idparam)
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
	server.serializeResponse(w, http.StatusOK, schedule)
	log.Print("Success! ", schedule.Scheduleid, " was received")
}

// TODO:Error handling and logs
func (server *Server) findallschedules(w http.ResponseWriter, r *http.Request) {
	page_id := r.URL.Query().Get("page_id")
	page_size := r.URL.Query().Get("page_size")
	pageid, _ := strconv.Atoi(page_id)
	if pageid < 1 {
		http.Error(w, "Page id can't be less than 1", http.StatusBadRequest)
		return
	}
	pagesize, _ := strconv.Atoi(page_size)
	skip := (pageid - 1) * pagesize
	args := models.ListSchedules{
		Limit:  pagesize,
		Offset: skip,
	}
	schedules, err := server.Services.ScheduleService.FindAll(args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	server.serializeResponse(w, http.StatusOK, schedules)
	log.Print("Success! ", len(schedules), " request")
}
