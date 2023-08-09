package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
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
	session, _ := server.Store.Get(r, "user-session")
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
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
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
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	http.Redirect(w, r, "/home", http.StatusMovedPermanently)
}

func (server *Server) home(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "user-session")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user := getUser(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}
	appointment, err := server.Services.AppointmentService.FindAllByPatient(user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	records, err := server.Services.PatientRecordService.FindAllByPatient(user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
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
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "home.html", data)
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
	var child bool
	Errmap := make(map[string]string)
	session, err := server.Store.Get(r, "user-session")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user := getUser(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}
	pat, err := server.Services.PatientService.Find(user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
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
		Success    string
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
	r.ParseMultipartForm(10 * 1024 * 1024)
	file, handler, err := r.FormFile("avatar")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["file"] = "avatar required"
		data.Errors = Errmap
		server.Templates.Render(w, "patient-profile.html", data)
		return
	}
	defer file.Close()
	avatar, err := server.UploadAvatar(file, strconv.Itoa(user.Id), "patient", handler.Filename)
	if err != nil {
		Errmap["file"] = err.Error()
		data.Errors = Errmap
		server.Templates.Render(w, "patient-profile.html", data)
		return
	}
	dob, _ := time.Parse("2006-01-02", register.Dob)
	hashed_password, _ := services.HashPassword(register.Password)
	if r.PostFormValue("Ischild") == "true" {
		child = true
	} else {
		child = false
	}
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
		Ischild:            child,
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
	w.WriteHeader(http.StatusOK)
	data.Patient = patient
	data.Success = "account updated successfully"
	server.Templates.Render(w, "patient-profile.html", data)
}

func (server *Server) PatientLogout(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "user-session")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	session.Values["user"] = PatientResp{}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	http.Redirect(w, r, "/home", http.StatusMovedPermanently)
}

func (server *Server) record(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "user-session")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user := getUser(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}
	w.WriteHeader(http.StatusOK)

	records, err := server.Services.PatientRecordService.FindAllByPatient(user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	data := struct {
		User    PatientResp
		Records []models.Patientrecords
	}{
		User:    user,
		Records: records,
	}
	server.Templates.Render(w, "records.html", data)
}

func (server *Server) appointments(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "user-session")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user := getUser(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}
	appointment, err := server.Services.AppointmentService.FindAllByPatient(user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	data := struct {
		User   PatientResp
		Apntmt []models.Appointment
	}{
		User:   user,
		Apntmt: appointment,
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "appointments.html", data)
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
	var child bool
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
	if r.PostFormValue("Ischild") == "true" {
		child = true
	} else {
		child = false
	}
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
		Ischild:         child,
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
	for i := 0; i < server.Worker.Nworker; i++ {
		go func() {
			defer server.Done()
			server.Worker.Workqueue()
		}()
	}
	http.Redirect(w, r, "/login", http.StatusMovedPermanently)
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
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "success.verify.html", nil)
}

func (server *Server) Patienteditappointment(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "user-session")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user := getUser(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	}
	appointment, err := server.Services.AppointmentService.FindAllByPatient(user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	data := struct {
		User   PatientResp
		Apntmt []models.Appointment
	}{
		User:   user,
		Apntmt: appointment,
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "appointments.html", data)
}

func (server *Server) Patientfilternurse(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "user-session")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	name := r.URL.Query().Get("name")
	user := getUser(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
	}
	form := NewForm(r, &Filter{})
	var ok = r.PostFormValue("Search")
	var filtermap = make(map[string]string)
	matches := matchsubstring(ok, keyvaluepairregex)
	for _, match := range matches {
		filtermap = filterkeypair(match[1], match[2], filtermap)
	}
	if len(filtermap) > 0 {
		if filtermap["name"] != "" {
			name = filtermap["name"]
			url := r.URL.Path + `?pageid=1` + "&" + "name=" + name
			http.Redirect(w, r, url, http.StatusMovedPermanently)
		}
	}
	id := r.URL.Query().Get("pageid")
	idparam, err := strconv.Atoi(id)
	if err != nil || idparam <= 0 {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	nurses, metadata, err := server.Services.NurseService.Filter(name, models.Filters{
		PageSize: PageCount,
		Page:     idparam,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	paging := Newpagination(*metadata)
	paging.nextpage(idparam)
	paging.previouspage(idparam)
	data := struct {
		User       PatientResp
		Nurses     []*models.Nurse
		Pagination Pagination
		Csrf       map[string]interface{}
	}{
		User:       user,
		Nurses:     nurses,
		Pagination: paging,
		Csrf:       form.Csrf}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "filter-nurse.html", data)
}
func (server *Server) PatienBookAppointment(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "nurse")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user := getNurse(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}
	var msg Form
	register := PatientAppointment{
		PatientEmail:    r.PostFormValue("Email"),
		AppointmentDate: r.PostFormValue("Appointmentdate"),
		Duration:        r.PostFormValue("Duration"),
	}
	msg = NewForm(r, &register)
	data := struct {
		User    NurseResp
		Errors  Errors
		Csrf    map[string]interface{}
		Success string
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
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	date, err := time.Parse("2006-01-02T15:04", r.PostFormValue("Appointmentdate"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	patient, err := server.Services.PatientService.FindbyEmail(register.PatientEmail)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Exists"] = err.Error()
		data.Errors = msg.Errors
		server.Templates.Render(w, "book-appointment.html", data)
		return
	}
	apntmt := models.Appointment{
		Doctorid:        doctorid,
		Patientid:       patient.Patientid,
		Appointmentdate: date,
		Duration:        register.Duration,
		Approval:        true,
	}
	_, err = server.Services.PatientBookAppointment(apntmt)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Exists"] = err.Error()
		data.Errors = msg.Errors
		server.Templates.Render(w, "book-appointment.html", data)
		return
	}
	w.WriteHeader(http.StatusCreated)
	data.Success = "appointment created successfully"
	server.Templates.Render(w, "book-appointment.html", data)
}
func (server *Server) PatientViewRecord(w http.ResponseWriter, r *http.Request) {
	errmap := make(map[string]string)
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		server.Templates.Render(w, "404.html", nil)
		return
	}
	data, err := server.Services.PatientRecordService.Find(idparam)
	if err != nil {
		server.Templates.Render(w, "404.html", nil)
	}
	session, err := server.Store.Get(r, "user-session")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user := getUser(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	}
	if user.Id != data.Patienid {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}
	pdata := struct {
		User    PatientResp
		Errors  Errors
		Records models.Patientrecords
	}{
		Errors:  errmap,
		Records: data,
		User:    user,
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "view-record.html", pdata)
}
func (server *Server) PatientUpdateAppointment(w http.ResponseWriter, r *http.Request) {
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
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user := getUser(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
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
	}
	msg = NewForm(r, &register)
	pdata := struct {
		User        PatientResp
		Errors      Errors
		Csrf        map[string]interface{}
		Appointment models.Appointment
		Success     string
	}{
		Errors:      Errmap,
		Appointment: data,
		User:        user,
		Csrf:        msg.Csrf,
	}
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
	doctorid, _ := strconv.Atoi(r.PostFormValue("Doctorid"))
	patientid, _ := strconv.Atoi(r.PostFormValue("Patientid"))
	date, err := time.Parse("2006-01-02T15:04", r.PostFormValue("Appointmentdate"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	apntmt := models.Appointment{
		Appointmentid:   data.Appointmentid,
		Doctorid:        doctorid,
		Patientid:       patientid,
		Appointmentdate: date,
		Duration:        register.Duration,
		Approval:        false,
	}

	if _, err := server.Services.UpdateappointmentbyPatient(apntmt); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		pdata.Errors = Errmap
		server.Templates.Render(w, "update-appointment.html", pdata)
		return
	}
	w.WriteHeader(http.StatusOK)
	pdata.Appointment = apntmt
	pdata.Success = "appointment updated successfully"
	server.Templates.Render(w, "update-appointment.html", pdata)
}

func (server *Server) patient_reset_password(w http.ResponseWriter, r *http.Request) {
	Errmap := make(map[string]string)
	id := r.URL.Query().Get("id")
	if !strings.Contains(id, "patient") {
		w.WriteHeader(http.StatusNotFound)
		server.Templates.Render(w, "404.html", nil)
		return
	}
	value, err := server.Redis.Get(server.Context, id).Result()
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		server.Templates.Render(w, "404.html", nil)
		return
	}
	pat, err := server.Services.PatientService.FindbyEmail(value)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		server.Templates.Render(w, "404.html", nil)
		return
	}
	register := ResetPassword{
		Email:           r.PostFormValue("Email"),
		Password:        r.PostFormValue("Password"),
		ConfirmPassword: r.PostFormValue("ConfirmPassword"),
	}
	msg := NewForm(r, &register)
	data := struct {
		Patient    models.Patient
		Errors     Errors
		Csrf       map[string]interface{}
		Bloodgroup []string
		Success    string
	}{
		Errors:     Errmap,
		Patient:    pat,
		Bloodgroup: bloodgroup_array(),
		Csrf:       msg.Csrf,
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "password_reset.html", data)
		return
	}
	if ok := msg.Validate(); !ok {
		data.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "password_reset.html", data)
		return
	}
	pat.Password_change_at = time.Now()
	pat.Hashed_password, err = services.HashPassword(register.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	if _, err := server.Services.PatientService.Update(pat); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		data.Errors = Errmap
		server.Templates.Render(w, "password_reset.html", data)
		return
	}
	w.WriteHeader(http.StatusOK)
	data.Success = "password reset successfully"
	server.Templates.Render(w, "password_reset.html", data)
	server.Redis.Del(server.Context, id)
}
func (server *Server) PatientTriage(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "user-session")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user := getUser(session)
	if !user.Authenticated {
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	}
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	var keyticket = "ticket" + utils.RandString(20)
	doctors, _, _ := server.Services.DoctorService.FindAll(models.Filters{Page: 1, PageSize: 20})
	patient, _ := server.Services.PatientService.Find(user.Id)
	if err = server.Redis.Set(server.Context, keyticket, Ticket{
		Ticketid:     keyticket,
		Patientemail: patient.Email,
		Doctorid:     doctors[0].Physicianid,
		Nurseid:      idparam,
		Attendedto:   false,
	}, 0).Err(); err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	http.Redirect(w, r, "/triages", http.StatusMovedPermanently)
}

func (server *Server) PatientListTriage(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "user-session")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user := getUser(session)
	if !user.Authenticated {
		http.Redirect(w, r, "/nurse/login", http.StatusMovedPermanently)
	}
	patient, _ := server.Services.PatientService.Find(user.Id)
	var tickets = server.getpatienttickets(patient.Email)
	data := struct {
		User    PatientResp
		Tickets []Ticket
	}{
		User:    user,
		Tickets: tickets,
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "patient-tickets.html", data)
}
