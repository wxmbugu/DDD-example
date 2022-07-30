package api

import (
	"fmt"
	"log"
	"net/http"
)

type DepartmentReq struct {
	Departmentid   int
	Departmentname string
}

func (server *Server) createdepartment(w http.ResponseWriter, r *http.Request) {
	var dep DepartmentReq
	err := decodejson(w, r, &dep)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	fmt.Println(dep)
	log.Println("Department created!")
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
