package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	"github.com/patienttracker/internal/models"
)

type DepartmentReq struct {
	Departmentname string `json:"departmentname" validate:"required"`
}

func (server *Server) createdepartment(w http.ResponseWriter, r *http.Request) {
	var dep DepartmentReq
	err := decodejson(w, r, &dep)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		server.Log.PrintError(err, fmt.Sprintf("Agent: %s, URL: %s", r.UserAgent(), r.URL.Path), fmt.Sprintf("ResponseCode:%d", http.StatusBadRequest))
		return
	}
	validate := validator.New()
	err = validate.Struct(dep)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		server.Log.PrintError(err, "some error happened!")
		return
	}
	department := models.Department{
		Departmentname: dep.Departmentname,
	}
	department, err = server.Services.DepartmentService.Create(department.Departmentname)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		server.Log.PrintError(err, fmt.Sprintf("Agent: %s, URL: %s", r.UserAgent(), r.URL.Path), fmt.Sprintf("ResponseCode:%d", http.StatusBadRequest))
		return
	}
	server.serializeResponse(w, http.StatusOK, department)
}

func (server *Server) updatedepartment(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	var dep DepartmentReq
	err = decodejson(w, r, &dep)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	validate := validator.New()
	err = validate.Struct(dep)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	department := models.Department{
		Departmentid:   idparam,
		Departmentname: dep.Departmentname,
	}
	department, err = server.Services.DepartmentService.Update(department.Departmentname, department.Departmentid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	server.serializeResponse(w, http.StatusOK, department)
	log.Print("Success! ", department.Departmentid, " was updated")
}

func (server *Server) deletedepartment(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	err = server.Services.DepartmentService.Delete(idparam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	server.serializeResponse(w, http.StatusOK, "department deleted successfully")
	log.Print("Success! ", idparam, " was deleted")
}

// TODO:Error handling and logs
func (server *Server) findalldepartment(w http.ResponseWriter, r *http.Request) {
	page_id := r.URL.Query().Get("page_id")
	page_size := r.URL.Query().Get("page_size")
	pageid, _ := strconv.Atoi(page_id)
	if pageid < 1 {
		http.Error(w, "Page id can't be less than 1", http.StatusBadRequest)
		return
	}
	pagesize, _ := strconv.Atoi(page_size)
	skip := (pageid - 1) * pagesize
	departments, err := server.Services.DepartmentService.FindAll(pagesize, skip)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error(), r.URL.Path, http.StatusBadRequest)
		return
	}
	server.serializeResponse(w, http.StatusOK, departments)
	log.Print("Success! ", len(departments), " request")
}
