package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/sessions"
	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/services"
)

type NurseResp struct {
	Id            int
	Username      string
	Email         string
	Authenticated bool
}

func getNurse(s *sessions.Session) NurseResp {
	val := s.Values["nurse"]
	var nurse = NurseResp{}
	nurse, ok := val.(NurseResp)
	if !ok {
		return NurseResp{Authenticated: false}
	}
	return nurse
}
func NurseResponse(nurse models.Nurse) NurseResp {
	return NurseResp{
		Username:      nurse.Username,
		Id:            nurse.Id,
		Authenticated: true,
	}
}
func (server *Server) NurseLogin(w http.ResponseWriter, r *http.Request) {
	var msg Form
	session, err := server.Store.Get(r, "nurse")
	if err = session.Save(r, w); err != nil {
		http.Redirect(w, r, "/500", 300)
	}
	login := Login{
		Email:    r.PostFormValue("email"),
		Password: r.PostFormValue("password"),
	}
	msg = NewForm(r, &login)

	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "staff-login.html", msg)
		return
	}
	if ok := msg.Validate(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "staff-login.html", msg)
		return
	}
	nurse, err := server.Services.NurseService.FindbyEmail(login.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusBadRequest)
			msg.Errors["Login"] = "No such user"
			server.Templates.Render(w, "staff-login.html", msg)
			return
		}
		http.Redirect(w, r, "/500", 300)
	}
	if err = services.CheckPassword(nurse.Hashed_password, login.Password); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Login"] = "No such user"
		server.Templates.Render(w, "staff-login.html", msg)
		return
	}
	user := NurseResponse(nurse)
	gobRegister(user)
	session.Values["nurse"] = user
	if err = session.Save(r, w); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		http.Redirect(w, r, "/500", 300)
	}
	http.Redirect(w, r, "/nurse/home", 300)
}

func (server *Server) NurseCreateRecord(w http.ResponseWriter, r *http.Request) {
	// BUG: A doctor who doesn't have an appointment with the said patient can create a record!!!!!
	// TODO: Might not Ideal but a fix woould to loop the appointments and check if there's an appointment with the said subjects
	session, err := server.Store.Get(r, "nurse")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	nurse := getNurse(session)
	if !nurse.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/staff/login", http.StatusSeeOther)
		return
	}
	var msg Form
	height, _ := strconv.Atoi(r.PostFormValue("Height"))
	bp, _ := strconv.Atoi(r.PostFormValue("Bp"))
	temp, _ := strconv.Atoi(r.PostFormValue("Temperature"))
	patientid, _ := strconv.Atoi(r.PostFormValue("Patientid"))
	_, _ = strconv.Atoi(r.PostFormValue("Doctorid"))
	register := NurseRecords{
		Height:      r.PostFormValue("Height"),
		Bp:          r.PostFormValue("Bp"),
		Temperature: r.PostFormValue("Temperature"),
		Weight:      r.PostFormValue("Weight"),
		Patientid:   r.PostFormValue("Patientid"),
		Doctorid:    r.PostFormValue("Doctorid"),
		// Additional:  r.PostFormValue("Additional"),
	}
	msg = NewForm(r, &register)
	data := struct {
		User   NurseResp
		Errors Errors
		Csrf   map[string]interface{}
	}{
		User:   nurse,
		Errors: msg.Errors,
		Csrf:   msg.Csrf,
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "nurse-edit-record.html", data)
		return
	}
	if ok := msg.Validate(); !ok {
		data.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "nurse-edit-record.html", data)
		return
	}
	records := models.Patientrecords{
		Patienid:    patientid,
		Nurseid:     nurse.Id,
		Height:      height,
		Bp:          bp,
		Temperature: temp,
		Weight:      register.Weight,
		Additional:  r.PostFormValue("Additional"),
		Date:        time.Now(),
	}
	if _, err := server.Services.PatientRecordService.Create(records); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Exists"] = "record already exist"
		data.Errors = msg.Errors
		server.Templates.Render(w, "nurse-edit-record.html", data)
		return
	}
	http.Redirect(w, r, "/nurse/home", 300)
}

func (server *Server) NurseLogout(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "nurse")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	session.Values["nurse"] = NurseResp{}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	http.Redirect(w, r, "/nurse/login", 300)
}
func (server *Server) Nurserecord(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "nurse")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	user := getNurse(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/staff/login", http.StatusSeeOther)
		return
	}
	w.WriteHeader(http.StatusOK)

	records, err := server.Services.PatientRecordService.FindAllByNurse(user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	data := struct {
		User    NurseResp
		Records []models.Patientrecords
	}{
		User:    user,
		Records: records,
	}
	server.Templates.Render(w, "nurse-records.html", data)
	return
}

// func (server *server) staffupdaterecord(w http.responsewriter, r *http.request) {
// 	var msg form
// 	errmap := make(map[string]string)
// 	params := mux.vars(r)
// 	id := params["id"]
// 	idparam, err := strconv.atoi(id)
// 	if err != nil {
// 		http.error(w, err.error(), http.statusbadrequest)
// 		return
// 	}
// 	data, err := server.services.patientrecordservice.find(idparam)
// 	if err != nil {
// 		server.templates.render(w, "404.html", nil)
// 	}
// 	session, err := server.store.get(r, "staff")
// 	if err != nil {
// 		w.writeheader(http.statusinternalservererror)
// 		http.redirect(w, r, "/500", 300)
// 	}
// 	user := getstaff(session)
// 	if !user.authenticated {
// 		w.writeheader(http.statusunauthorized)
// 		http.redirect(w, r, "/staff/login", http.statusseeother)
// 		return
// 	}
// 	if user.id != data.doctorid {
// 		w.writeheader(http.statusunauthorized)
// 		server.templates.render(w, "401.html", nil)
// 		return
// 	}
// 	pdata := struct {
// 		user    doctorresp
// 		errors  errors
// 		records models.patientrecords
// 		csrf    map[string]interface{}
// 	}{
// 		errors:  errmap,
// 		records: data,
// 		user:    user,
// 		csrf:    msg.csrf,
// 	}
// 	// var approval bool
// 	register := records{
// 		patientid:    r.postformvalue("doctorid"),
// 		doctorid:     r.postformvalue("doctorid"),
// 		diagnosis:    r.postformvalue("diagnosis"),
// 		disease:      r.postformvalue("disease"),
// 		prescription: r.postformvalue("prescription"),
// 		weight:       r.postformvalue("weight"),
// 	}
// 	msg = newform(r, &register)
// 	if r.method == "get" {
// 		w.writeheader(http.statusok)
// 		server.templates.render(w, "staff-update-record.html", pdata)
// 		return
// 	}
// 	if ok := msg.validate(); !ok {
// 		pdata.errors = msg.errors
// 		w.writeheader(http.statusbadrequest)
// 		server.templates.render(w, "staff-update-record.html", pdata)
// 		return
// 	}
//
// 	dt := struct {
// 		user   doctorresp
// 		errors errors
// 		csrf   map[string]interface{}
// 	}{
// 		user:   user,
// 		csrf:   msg.csrf,
// 		errors: errmap,
// 	}
// 	doctorid, _ := strconv.atoi(r.postformvalue("doctorid"))
// 	patientid, _ := strconv.atoi(r.postformvalue("patientid"))
// 	records := models.patientrecords{
// 		recordid:     data.recordid,
// 		patienid:     patientid,
// 		doctorid:     doctorid,
// 		diagnosis:    register.diagnosis,
// 		disease:      register.diagnosis,
// 		prescription: register.prescription,
// 		weight:       register.weight,
// 		date:         data.date,
// 	}
//
// 	if _, err := server.services.patientrecordservice.update(records); err != nil {
// 		w.writeheader(http.statusbadrequest)
// 		errmap["exists"] = err.error()
// 		dt.errors = errmap
// 		server.templates.render(w, "staff-update-record.html", dt)
// 		return
// 	}
// 	http.redirect(w, r, r.url.string(), 301)
// }
