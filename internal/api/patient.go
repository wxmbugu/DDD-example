package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/services"
	"github.com/patienttracker/internal/utils"
)

type PatientResp struct {
	Id            int
	Username      string
	Full_name     string
	Avatar        string
	Authenticated bool
}

//TODO: set env of tokenduration

const tokenduration = 45

func PatientResponse(patient models.Patient) PatientResp {
	return PatientResp{
		Username:      patient.Username,
		Full_name:     patient.Full_name,
		Id:            patient.Patientid,
		Avatar:        patient.Avatar,
		Authenticated: true,
	}
}

func (server *Server) PatientLogin(w http.ResponseWriter, r *http.Request) {
	var msg Form
	login := Login{
		Email:    r.PostFormValue("email"),
		Password: r.PostFormValue("password"),
	}
	msg = NewForm(r, &login)
	session, err := server.Store.Get(r, "user-session")
	if err = session.Save(r, w); err != nil {
		http.Redirect(w, r, "/500", 300)

	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "login.html", msg)
		return
	}
	if ok := msg.Validate(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "login.html", msg)
		return
	}
	patient, err := server.Services.PatientService.FindbyEmail(login.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusBadRequest)
			msg.Errors["Login"] = "No such user"
			server.Templates.Render(w, "login.html", msg)
			return
		}
		http.Redirect(w, r, "/500", 300)
	}
	if err = services.CheckPassword(patient.Hashed_password, login.Password); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Login"] = "No such user"
		server.Templates.Render(w, "login.html", msg)
		return
	}
	user := PatientResponse(patient)
	gobRegister(user)
	session.Values["user"] = user
	if err = session.Save(r, w); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		http.Redirect(w, r, "/500", 300)
	}
	http.Redirect(w, r, "/home", 300)
}

func (server *Server) home(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "user-session")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	user := getUser(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	w.WriteHeader(http.StatusOK)

	appointment, err := server.Services.AppointmentService.FindAllByPatient(user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	records, err := server.Services.PatientRecordService.FindAllByPatient(user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	data := struct {
		User    PatientResp
		Apntmt  []models.Appointment
		Records []models.Patientrecords
	}{
		User:    user,
		Apntmt:  appointment,
		Records: records,
	}
	server.Templates.Render(w, "home.html", data)
	return

}

func bloodgroup_array() []string {
	var bloodgroup = []string{
		"A+",
		"A-",
		"B+",
		"B-",
		"AB+",
		"AB-",
		"O+",
		"O-",
	}
	return bloodgroup
}

func (server *Server) profile(w http.ResponseWriter, r *http.Request) {
	Errmap := make(map[string]string)
	session, err := server.Store.Get(r, "user-session")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 301)
	}
	user := getUser(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	pat, err := server.Services.PatientService.Find(user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 301)
	}
	register := Register{
		Email:           r.PostFormValue("Email"),
		Password:        r.PostFormValue("Password"),
		ConfirmPassword: r.PostFormValue("ConfirmPassword"),
		Username:        r.PostFormValue("Username"),
		Fullname:        r.PostFormValue("Fullname"),
		Contact:         r.PostFormValue("Contact"),
		Dob:             r.PostFormValue("Dob"),
		Bloodgroup:      r.PostFormValue("Bloodgroup"),
	}
	msg := NewForm(r, &register)
	data := struct {
		User       PatientResp
		Patient    models.Patient
		Errors     Errors
		Csrf       map[string]interface{}
		Bloodgroup []string
	}{
		User:       user,
		Patient:    pat,
		Bloodgroup: bloodgroup_array(),
		Csrf:       msg.Csrf,
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "patient-profile.html", data)
		return
	}
	if ok := msg.Validate(); !ok {
		data.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "patient-profile.html", data)
		return
	}
	file, handler, err := r.FormFile("avatar")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["file"] = "avatar required"
		data.Errors = Errmap
		server.Templates.Render(w, "patient-profile.html", data)
		return

	}
	defer file.Close()
	if handler.Size > 20*1024*1024 {
		Errmap["size"] = "file is larger than 20mb"
		data.Errors = Errmap
		server.Templates.Render(w, "patient-profile.html", data)
		return
	}
	avatar, err := server.UploadAvatar(file, strconv.Itoa(user.Id), "staff", handler.Filename)
	if err != nil {
		Errmap["file"] = err.Error()
		data.Errors = Errmap
		server.Templates.Render(w, "patient-profile.html", data)
		return
	}
	dob, _ := time.Parse("2006-01-02", register.Dob)
	hashed_password, _ := services.HashPassword(register.Password)
	patient := models.Patient{
		Patientid:          user.Id,
		Username:           register.Username,
		Full_name:          register.Fullname,
		Email:              register.Email,
		Dob:                dob,
		Contact:            register.Contact,
		Avatar:             avatar,
		Bloodgroup:         register.Bloodgroup,
		Hashed_password:    hashed_password,
		Verified:           false,
		About:              r.PostFormValue("About"),
		Password_change_at: time.Now(),
	}
	if _, err := server.Services.PatientService.Update(patient); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		data.Errors = Errmap
		server.Templates.Render(w, "patient-profile.html", data)
		return
	}
	http.Redirect(w, r, r.URL.String(), 301)
}

func (server *Server) PatientLogout(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "user-session")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	session.Values["user"] = PatientResp{}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	http.Redirect(w, r, "/home", 300)
}

func (server *Server) record(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "user-session")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	user := getUser(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	w.WriteHeader(http.StatusOK)

	records, err := server.Services.PatientRecordService.FindAllByPatient(user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	data := struct {
		User PatientResp
		// Apntmt  []models.Appointment
		Records []models.Patientrecords
	}{
		User: user,
		// Apntmt:  appointment,
		Records: records,
	}
	server.Templates.Render(w, "records.html", data)
	return

}

func (server *Server) appointments(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "user-session")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	user := getUser(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	w.WriteHeader(http.StatusOK)

	appointment, err := server.Services.AppointmentService.FindAllByPatient(user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	data := struct {
		User   PatientResp
		Apntmt []models.Appointment
	}{
		User:   user,
		Apntmt: appointment,
	}
	server.Templates.Render(w, "appointments.html", data)
	return

}
func getUser(s *sessions.Session) PatientResp {
	val := s.Values["user"]
	var user = PatientResp{}
	user, ok := val.(PatientResp)
	if !ok {
		return PatientResp{Authenticated: false}
	}
	return user
}

func (server *Server) createpatient(w http.ResponseWriter, r *http.Request) {
	var msg Form
	register := Register{
		Email:           r.PostFormValue("Email"),
		Password:        r.PostFormValue("Password"),
		ConfirmPassword: r.PostFormValue("ConfirmPassword"),
		Username:        r.PostFormValue("Username"),
		Fullname:        r.PostFormValue("Fullname"),
		Contact:         r.PostFormValue("Contact"),
		Dob:             r.PostFormValue("Dob"),
		Bloodgroup:      r.PostFormValue("Bloodgroup"),
	}
	msg = NewForm(r, &register)
	var dataform = struct {
		Msg        Form
		Errors     map[string]string
		Bloodgroup []string
	}{
		Msg:        msg,
		Errors:     msg.Errors,
		Bloodgroup: bloodgroup_array(),
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "register.html", dataform)
		return
	}
	if ok := msg.Validate(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		dataform.Errors = msg.Errors
		server.Templates.Render(w, "register.html", dataform)
		return
	}
	dob, _ := time.Parse("2006-01-02", register.Dob)

	hashed_password, _ := services.HashPassword(register.Password)
	patient := models.Patient{
		Username:        register.Username,
		Full_name:       register.Fullname,
		Email:           register.Email,
		Dob:             dob,
		Contact:         register.Contact,
		Bloodgroup:      register.Bloodgroup,
		Hashed_password: hashed_password,
		Verified:        false,
		Created_at:      time.Now(),
	}
	if _, err := server.Services.PatientService.Create(patient); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Exists"] = "User already Exists"
		dataform.Errors = msg.Errors
		server.Templates.Render(w, "register.html", dataform)
		return
	}
	key := utils.RandString(20)
	value := patient.Email
	err := server.Redis.Set(server.Context, key, value, 0).Err()
	if err != nil {
		server.Log.Error(err)
	}
	data := struct {
		URL   string
		Name  string
		Email string
	}{
		URL:   `http://localhost:9000/verify/` + key,
		Name:  patient.Username,
		Email: patient.Email,
	}
	mailer := server.Mailer.setdata(data, "Welcome to Our System!!", "verify.account.html", data.Email)
	go func() {
		server.Worker.Task <- &mailer
	}()
	server.WaitGroup.Add(server.Worker.Nworker / 2)
	for i := 0; i < (server.Worker.Nworker)/2; i++ {
		go func() {
			defer server.Done()
			server.Worker.Workqueue()
		}()
	}
	http.Redirect(w, r, "/login", 300)
}
func (server *Server) VerifyAccount(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	value, err := server.Redis.Get(server.Context, id).Result()
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		server.Templates.Render(w, "404.html", nil)
	}
	data, err := server.Services.PatientService.FindbyEmail(value)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		server.Templates.Render(w, "404.html", nil)
	}
	data.Verified = true
	server.Services.PatientService.Update(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "success.verify.html", nil)
	return

}

//	func (server *Server) background(fn func()) {
//		server.Wg.Add(1)
//		go func() {
//			defer server.Wg.Done()
//			defer func() {
//				if err := recover(); err != nil {
//					server.Log.Error(errors.New(err.(string)), nil)
//				}
//			}()
//			fn()
//		}()
//	}
//
// TODO: Paiient Edit Appointment
func (server *Server) Patienteditappointment(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "user-session")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	user := getUser(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	w.WriteHeader(http.StatusOK)

	appointment, err := server.Services.AppointmentService.FindAllByPatient(user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	data := struct {
		User   PatientResp
		Apntmt []models.Appointment
	}{
		User:   user,
		Apntmt: appointment,
	}
	server.Templates.Render(w, "appointments.html", data)
	return

}

func (server *Server) Patientshowdepartments(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "user-session")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	user := getUser(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	departments, err := server.Services.DepartmentService.FindAll(models.ListDepartment{
		Limit:  10000,
		Offset: 0,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	data := struct {
		User       PatientResp
		Department []models.Department
	}{
		User:       user,
		Department: departments,
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "department.html", data)
	return
}

func (server *Server) PatientListDoctorsDept(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "user-session")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	user := getUser(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	w.WriteHeader(http.StatusOK)
	params := mux.Vars(r)
	deptname := params["name"]
	doctors, err := server.Services.DoctorService.FindDoctorsbyDept(models.ListDoctorsbyDeptarment{
		Department: deptname,
		Limit:      100000,
		Offset:     0,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	data := struct {
		User    PatientResp
		Doctors []models.Physician
	}{
		User:    user,
		Doctors: doctors,
	}
	server.Templates.Render(w, "department-doctors.html", data)
	return
}

func (server *Server) PatienBookAppointment(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "user-session")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	user := getUser(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	var msg Form
	register := PatientAppointment{
		AppointmentDate: r.PostFormValue("Appointmentdate"),
		Duration:        r.PostFormValue("Duration"),
	}
	msg = NewForm(r, &register)
	data := struct {
		User   PatientResp
		Errors Errors
		Csrf   map[string]interface{}
	}{
		User:   user,
		Errors: msg.Errors,
		Csrf:   msg.Csrf,
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "book-appointment.html", data)
		return
	}
	if ok := msg.Validate(); !ok {
		data.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "book-appointment.html", data)
		return
	}
	params := mux.Vars(r)
	id := params["id"]
	doctorid, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	date, err := time.Parse("2006-01-02T15:04", r.PostFormValue("Appointmentdate"))
	apntmt := models.Appointment{
		Doctorid:        doctorid,
		Patientid:       user.Id,
		Appointmentdate: date,
		Duration:        register.Duration,
		Approval:        false,
	}
	_, err = server.Services.PatientBookAppointment(apntmt)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Exists"] = err.Error()
		data.Errors = msg.Errors
		server.Templates.Render(w, "book-appointment.html", data)
		return
	}
	http.Redirect(w, r, "/appointments", 300)

}
func (server *Server) PatienUpdateAppointment(w http.ResponseWriter, r *http.Request) {
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
		server.Templates.Render(w, "404.html", nil)
	}
	session, err := server.Store.Get(r, "user-session")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	user := getUser(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if user.Id != data.Patientid {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return

	}
	register := Appointment{
		Doctorid:        r.PostFormValue("Doctorid"),
		Patientid:       r.PostFormValue("Patientid"),
		AppointmentDate: r.PostFormValue("Appointmentdate"),
		Duration:        r.PostFormValue("Duration"),
		Approval:        r.PostFormValue("Approval"),
	}
	msg = NewForm(r, &register)
	pdata := struct {
		User        PatientResp
		Errors      Errors
		Csrf        map[string]interface{}
		Appointment models.Appointment
	}{
		Errors:      Errmap,
		Appointment: data,
		User:        user,
		Csrf:        msg.Csrf,
	}
	var approval bool
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "update-appointment.html", pdata)
		return
	}
	if ok := msg.Validate(); !ok {
		pdata.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "update-appointment.html", pdata)
		return
	}

	dt := struct {
		User   PatientResp
		Errors Errors
		Csrf   map[string]interface{}
	}{
		User:   user,
		Errors: Errmap,
		Csrf:   msg.Csrf,
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

	if _, err := server.Services.UpdateappointmentbyPatient(apntmt); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		dt.Errors = Errmap
		server.Templates.Render(w, "update-appointment.html", dt)
		return
	}
	http.Redirect(w, r, r.URL.String(), 301)
}
