package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/patienttracker/internal/models"
)

type DepartmentReq struct {
	Departmentid   int    `json:"departmentid" validate:"required"`
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
		Departmentid:   dep.Departmentid,
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
	log.Println("Department updated")
}

func (server *Server) deletedepartment(w http.ResponseWriter, r *http.Request) {
	log.Println("Department deleted!")
}

func (server *Server) findalldepartment(w http.ResponseWriter, r *http.Request) {
	log.Println("Department found!")
}
