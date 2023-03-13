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
	"github.com/gorilla/sessions"
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

type DoctorResp struct {
	Id            int
	Username      string
	Email         string
	Authenticated bool
}

func DoctorResponse(doctor models.Physician) DoctorResp {
	return DoctorResp{
		Id:            doctor.Physicianid,
		Username:      doctor.Username,
		Email:         doctor.Email,
		Authenticated: true,
	}
}
func getStaff(s *sessions.Session) DoctorResp {
	val := s.Values["staff"]
	var staff = DoctorResp{}
	staff, ok := val.(DoctorResp)
	if !ok {
		return DoctorResp{Authenticated: false}
	}
	return staff
}
func (server *Server) StaffLogin(w http.ResponseWriter, r *http.Request) {
	var msg Form
	session, err := server.Store.Get(r, "staff")
	if err = session.Save(r, w); err != nil {
		http.Redirect(w, r, "/500", 300)

	}
	login := Login{
		Email:    r.PostFormValue("email"),
		Password: r.PostFormValue("password"),
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "staff-login.html", msg)
		return
	}
	msg = Form{
		Data: &login,
	}
	if ok := msg.Validate(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "staff-login.html", msg)
		return
	}
	user, err := server.Services.DoctorService.FindbyEmail(login.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusBadRequest)
			msg.Errors["Login"] = "No such user"
			server.Templates.Render(w, "staff-login.html", msg)
			return
		}
		http.Redirect(w, r, "/500", 300)
	}
	if err = services.CheckPassword(user.Hashed_password, login.Password); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Login"] = "No such user"
		server.Templates.Render(w, "staff-login.html", msg)
		return
	}

	staff := DoctorResponse(user)
	gobRegister(staff)
	session.Values["staff"] = staff
	if err = session.Save(r, w); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		http.Redirect(w, r, "/500", 300)
	}
	http.Redirect(w, r, "/staff/home", http.StatusSeeOther)
}
func (server *Server) Staffprofile(w http.ResponseWriter, r *http.Request) {
	Errmap := make(map[string]string)
	session, err := server.Store.Get(r, "staff")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	user := getStaff(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	doc, err := server.Services.DoctorService.Find(user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	register := DocRegister{
		Email:           r.PostFormValue("Email"),
		Password:        r.PostFormValue("Password"),
		ConfirmPassword: r.PostFormValue("ConfirmPassword"),
		Username:        r.PostFormValue("Username"),
		Fullname:        r.PostFormValue("Fullname"),
		Contact:         r.PostFormValue("Contact"),
		Departmentname:  r.PostFormValue("Departmentname"),
	}
	msg := Form{
		Data: &register,
	}
	data := struct {
		User   DoctorResp
		Doctor models.Physician
		Errors Errors
	}{
		User:   user,
		Doctor: doc,
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "doctor-profile.html", data)
		return
	}
	if ok := msg.Validate(); !ok {
		data.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "doctor-profile.html", data)
		return
	}

	hashed_password, _ := services.HashPassword(register.Password)
	doctor := models.Physician{
		Physicianid:         user.Id,
		Username:            register.Username,
		Full_name:           register.Fullname,
		Email:               register.Email,
		Contact:             register.Contact,
		About:               r.PostFormValue("About"),
		Verified:            false,
		Hashed_password:     hashed_password,
		Departmentname:      register.Departmentname,
		Password_changed_at: time.Now(),
	}
	newdoctor, err := server.Services.DoctorService.Update(doctor)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		data.Errors = Errmap
		server.Templates.Render(w, "doctor-profile.html", data)
		return
	}
	data.Errors = msg.Errors
	data.Doctor = newdoctor
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "doctor-profile.html", data)
}
func (server *Server) Staffhome(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "staff")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	user := getStaff(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/staff/login", http.StatusSeeOther)
		return
	}
	appointment, err := server.Services.AppointmentService.FindAllByDoctor(user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	records, err := server.Services.PatientRecordService.FindAllByDoctor(user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	data := struct {
		User    DoctorResp
		Apntmt  []models.Appointment
		Records []models.Patientrecords
	}{
		User:    user,
		Apntmt:  appointment,
		Records: records,
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "staff-home.html", data)
	return
}

func (server *Server) Staffappointments(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "staff")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	user := getStaff(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/staff/login", http.StatusSeeOther)
		return
	}
	w.WriteHeader(http.StatusOK)

	appointment, err := server.Services.AppointmentService.FindAllByDoctor(user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	data := struct {
		User   DoctorResp
		Apntmt []models.Appointment
	}{
		User:   user,
		Apntmt: appointment,
	}
	server.Templates.Render(w, "staff-appointments.html", data)
	return

}
func (server *Server) Staffrecord(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "staff")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	user := getStaff(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/staff/login", http.StatusSeeOther)
		return
	}
	w.WriteHeader(http.StatusOK)

	records, err := server.Services.PatientRecordService.FindAllByDoctor(user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	data := struct {
		User    DoctorResp
		Records []models.Patientrecords
	}{
		User:    user,
		Records: records,
	}
	server.Templates.Render(w, "staff-records.html", data)
	return
}

func (server *Server) StaffUpdateAppointment(w http.ResponseWriter, r *http.Request) {
	var msg Form
	Errmap := make(map[string]string)
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := server.Services.AppointmentService.Find(idparam)
	if err != nil {
		server.Templates.Render(w, "staff-update-appointment.html", "Schedule not found")
	}
	session, err := server.Store.Get(r, "staff")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	user := getStaff(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/staff/login", http.StatusSeeOther)
		return
	}
	register := Appointment{
		Doctorid:        r.PostFormValue("Doctorid"),
		Patientid:       r.PostFormValue("Patientid"),
		AppointmentDate: r.PostFormValue("Appointmentdate"),
		Duration:        r.PostFormValue("Duration"),
		Approval:        r.PostFormValue("Approval"),
	}
	msg = Form{
		Data: &register,
	}
	pdata := struct {
		User        DoctorResp
		Errors      Errors
		Appointment models.Appointment
	}{
		Errors:      Errmap,
		Appointment: data,
		User:        user,
	}
	var approval bool
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "staff-update-appointment.html", pdata)
		return
	}
	if ok := msg.Validate(); !ok {
		pdata.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "staff-update-appointment.html", pdata)
		return
	}

	dt := struct {
		User   DoctorResp
		Errors Errors
	}{
		User:   user,
		Errors: Errmap,
	}
	doctorid, _ := strconv.Atoi(r.PostFormValue("Doctorid"))
	patientid, _ := strconv.Atoi(r.PostFormValue("Patientid"))
	date, err := time.Parse("2006-01-02T15:04", r.PostFormValue("Appointmentdate"))
	if r.PostFormValue("Approval") == "Active" {
		approval = true
	} else if r.PostFormValue("Approval") == "Inactive" {
		approval = false
	} else {
		msg.Errors["ApprovalInput"] = "Should be either Active or Inactive"
	}

	apntmt := models.Appointment{
		Appointmentid:   data.Appointmentid,
		Doctorid:        doctorid,
		Patientid:       patientid,
		Appointmentdate: date,
		Duration:        register.Duration,
		Approval:        approval,
	}

	if _, err := server.Services.UpdateappointmentbyPatient(apntmt.Patientid, apntmt); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		dt.Errors = Errmap
		server.Templates.Render(w, "staff-update-appointment.html", dt)
		return
	}
	http.Redirect(w, r, "/staff/appointments", 300)
}
func (server *Server) StaffCreateRecord(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "staff")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	staff := getStaff(session)
	if !staff.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/staff/login", http.StatusSeeOther)
		return
	}
	var msg Form
	params := mux.Vars(r)
	id := params["id"]
	patientid, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	register := StaffRecords{
		Diagnosis:    r.PostFormValue("Diagnosis"),
		Disease:      r.PostFormValue("Disease"),
		Prescription: r.PostFormValue("Prescription"),
		Weight:       r.PostFormValue("Weight"),
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "staff-edit-records.html", nil)
		return
	}
	msg = Form{
		Data: &register,
	}
	data := struct {
		User   DoctorResp
		Errors Errors
	}{
		User:   staff,
		Errors: msg.Errors,
	}
	if ok := msg.Validate(); !ok {
		data.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "staff-edit-records.html", data)
		return
	}
	records := models.Patientrecords{
		Patienid:     patientid,
		Doctorid:     staff.Id,
		Diagnosis:    register.Diagnosis,
		Disease:      register.Diagnosis,
		Prescription: register.Prescription,
		Weight:       register.Weight,
		Date:         time.Now(),
	}
	if _, err := server.Services.PatientRecordService.Create(records); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Exists"] = "record already exist"
		data.Errors = msg.Errors
		server.Templates.Render(w, "staff-edit-records.html", data)
		return
	}
	http.Redirect(w, r, "/staff/records", 300)
}

func (server *Server) StaffUpdateRecord(w http.ResponseWriter, r *http.Request) {
	var msg Form
	Errmap := make(map[string]string)
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := server.Services.PatientRecordService.Find(idparam)
	if err != nil {
		server.Templates.Render(w, "staff-update-record.html", "Schedule not found")
	}
	session, err := server.Store.Get(r, "staff")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	user := getStaff(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/staff/login", http.StatusSeeOther)
		return
	}

	pdata := struct {
		User    DoctorResp
		Errors  Errors
		Records models.Patientrecords
	}{
		Errors:  Errmap,
		Records: data,
		User:    user,
	}
	// var approval bool
	register := Records{
		Patientid:    r.PostFormValue("Doctorid"),
		Doctorid:     r.PostFormValue("Doctorid"),
		Diagnosis:    r.PostFormValue("Diagnosis"),
		Disease:      r.PostFormValue("Disease"),
		Prescription: r.PostFormValue("Prescription"),
		Weight:       r.PostFormValue("Weight"),
	}
	msg = Form{
		Data: &register,
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "staff-update-record.html", pdata)
		return
	}
	if ok := msg.Validate(); !ok {
		pdata.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "staff-update-record.html", pdata)
		return
	}

	dt := struct {
		User   DoctorResp
		Errors Errors
	}{
		User:   user,
		Errors: Errmap,
	}
	doctorid, _ := strconv.Atoi(r.PostFormValue("Doctorid"))
	patientid, _ := strconv.Atoi(r.PostFormValue("Patientid"))
	records := models.Patientrecords{
		Recordid:     data.Recordid,
		Patienid:     patientid,
		Doctorid:     doctorid,
		Diagnosis:    register.Diagnosis,
		Disease:      register.Diagnosis,
		Prescription: register.Prescription,
		Weight:       register.Weight,
		Date:         data.Date,
	}

	if _, err := server.Services.PatientRecordService.Update(records); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		dt.Errors = Errmap
		server.Templates.Render(w, "staff-update-record.html", dt)
		return
	}
	http.Redirect(w, r, "/staff/records", 300)
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
