package api

import (
	"database/sql"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	// "github.com/go-acme/lego/v4/log"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/services"
)

type DoctorResp struct {
	Id            int
	Username      string
	Email         string
	Avatar        string
	Authenticated bool
}

func DoctorResponse(doctor models.Physician) DoctorResp {
	return DoctorResp{
		Id:            doctor.Physicianid,
		Username:      doctor.Username,
		Email:         doctor.Email,
		Avatar:        doctor.Avatar,
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
func (server *Server) Staffcreateschedule(w http.ResponseWriter, r *http.Request) {
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
	var msg Form
	var actvie bool
	register := Schedule{
		Doctorid:  r.PostFormValue("Doctorid"),
		Starttime: r.PostFormValue("Starttime"),
		Endtime:   r.PostFormValue("Endtime"),
		Active:    r.PostFormValue("Active"),
	}
	msg = NewForm(r, &register)
	data := struct {
		User   DoctorResp
		Active []string
		Errors Errors
		Csrf   map[string]interface{}
	}{
		User:   user,
		Active: active_inactive(),
		Errors: msg.Errors,
		Csrf:   msg.Csrf,
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "staff-edit-schedule.html", data)
		return
	}
	if ok := msg.Validate(); !ok {
		data.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "staff-edit-schedule.html", data)
		return
	}
	doctorid, _ := strconv.Atoi(r.PostFormValue("Doctorid"))
	if r.PostFormValue("Active") == "Active" {
		actvie = true
	} else if r.PostFormValue("Active") == "Inactive" {
		actvie = false
	} else {
		msg.Errors["AtiveInput"] = "Should be either Active or Inactive"
	}
	schedule := models.Schedule{
		Doctorid:  doctorid,
		Starttime: register.Starttime,
		Endtime:   register.Endtime,
		Active:    actvie,
	}
	if _, err := server.Services.MakeSchedule(schedule); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Exists"] = err.Error()
		data.Errors = msg.Errors
		server.Templates.Render(w, "staff-edit-schedule.html", data)
		return
	}
	http.Redirect(w, r, "/staff/home", 300)
}
func (server *Server) StaffLogout(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "staff")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	session.Values["staff"] = DoctorResp{}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	http.Redirect(w, r, "/staff/home", 300)
}
func (server *Server) Staffupdateschedule(w http.ResponseWriter, r *http.Request) {
	var msg Form
	Errmap := make(map[string]string)
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := server.Services.ScheduleService.Find(idparam)
	if err != nil {
		server.Templates.Render(w, "404.html", nil)
	}
	session, err := server.Store.Get(r, "staff")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	user := getStaff(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	register := Schedule{
		Doctorid:  r.PostFormValue("Doctorid"),
		Starttime: r.PostFormValue("Starttime"),
		Endtime:   r.PostFormValue("Endtime"),
		Active:    r.PostFormValue("Active"),
	}
	msg = NewForm(r, &register)
	pdata := struct {
		User     DoctorResp
		Errors   Errors
		Csrf     map[string]interface{}
		Schedule models.Schedule
		Active   []string
	}{
		Errors:   Errmap,
		Schedule: data,
		Csrf:     msg.Csrf,
		User:     user,
		Active:   active_inactive(),
	}
	var actvie bool
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "staff-update-schedule.html", pdata)
		return
	}
	if ok := msg.Validate(); !ok {
		pdata.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "staff-update-schedule.html", pdata)
		return
	}
	doctorid, _ := strconv.Atoi(r.PostFormValue("Doctorid"))
	if r.PostFormValue("Active") == "Active" {
		actvie = true
	} else if r.PostFormValue("Active") == "Inactive" {
		actvie = false
	} else {
		pdata.Errors["AtiveInput"] = "Should be either Active or Inactive"
	}
	dt := struct {
		User   DoctorResp
		Csrf   map[string]interface{}
		Errors Errors
		Active []string
	}{
		User:   user,
		Errors: Errmap,
		Active: active_inactive(),
		Csrf:   msg.Csrf,
	}
	schedule := models.Schedule{
		Scheduleid: data.Scheduleid,
		Doctorid:   doctorid,
		Starttime:  register.Starttime,
		Endtime:    register.Endtime,
		Active:     actvie,
	}
	if _, err := server.Services.UpdateSchedule(schedule); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		dt.Errors = Errmap
		server.Templates.Render(w, "staff-update-schedule.html", dt)
		return
	}
	http.Redirect(w, r, r.URL.String(), 301)
}
func (server *Server) Staffschedule(w http.ResponseWriter, r *http.Request) {
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

	schedules, err := server.Services.ScheduleService.FindbyDoctor(user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	data := struct {
		User     DoctorResp
		Schedule []models.Schedule
	}{
		User:     user,
		Schedule: schedules,
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "staff-schedule.html", data)
	return

}

func (server *Server) Staffdeleteschedule(w http.ResponseWriter, r *http.Request) {
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
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := server.Services.ScheduleService.Delete(idparam); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "staff-schedule.html", nil)
		return
	}
	http.Redirect(w, r, "/staff/schedules", 300)
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
	msg := NewForm(r, &register)
	data := struct {
		User   DoctorResp
		Doctor models.Physician
		Errors Errors
		Csrf   map[string]interface{}
	}{
		User:   user,
		Doctor: doc,
		Csrf:   msg.Csrf,
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
	r.ParseMultipartForm(10 * 1024 * 1024)
	file, handler, err := r.FormFile("avatar")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["file"] = "avatar required"
		data.Errors = Errmap
		server.Templates.Render(w, "doctor-profile.html", data)
		return
	}
	var avatar string
	defer file.Close()
	avatar, err = server.UploadAvatar(file, strconv.Itoa(user.Id), "staff", handler.Filename)
	if err != nil {
		Errmap["file"] = err.Error()
		data.Errors = Errmap
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
		Avatar:              avatar,
		Hashed_password:     hashed_password,
		Departmentname:      register.Departmentname,
		Password_changed_at: time.Now(),
	}
	if _, err := server.Services.DoctorService.Update(doctor); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		data.Errors = Errmap
		server.Templates.Render(w, "doctor-profile.html", data)
		return
	}
	http.Redirect(w, r, r.URL.String(), 301)
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

// TODO: Work on time parsing to check if it's 1h
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
		server.Templates.Render(w, "404.html", nil)
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
	if user.Id != data.Doctorid {
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
		User        DoctorResp
		Errors      Errors
		Appointment models.Appointment
		Csrf        map[string]interface{}
	}{
		Errors:      Errmap,
		Appointment: data,
		User:        user,
		Csrf:        msg.Csrf,
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

	appointment, err := server.Services.UpdateappointmentbyPatient(apntmt)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		dt.Errors = Errmap
		server.Templates.Render(w, "staff-update-appointment.html", dt)
		return
	}
	if appointment.Approval {
		// key := utils.RandString(20)
		err := server.Redis.Set(server.Context, strconv.Itoa(appointment.Appointmentid), appointment.Appointmentdate, 0).Err()
		if err != nil {
			server.Log.Error(err)
		}
	}
	http.Redirect(w, r, r.URL.String(), 301)
}

// AppointmentsSubscriber will be used to send upcoming appointments to our users via email
func (server *Server) AppointmentsEmailSender() {
	var ids []int
	var upcoming_appointment []models.Appointment
	iter := server.Redis.Scan(server.Context, 0, "*", 0).Iterator()
	for iter.Next(server.Context) {
		if int, err := strconv.Atoi(iter.Val()); err == nil {
			ids = append(ids, int)
		}
	}
	for _, v := range ids {
		appointment, err := server.Services.AppointmentService.Find(v)
		if err != nil {
			if err == sql.ErrNoRows {
				// delete appointments aren't in our database from redis
				server.Redis.Del(server.Context, strconv.Itoa(v))
			}
		}
		if int(appointment.Appointmentdate.Sub(time.Now()).Hours()) <= 24 {
			upcoming_appointment = append(upcoming_appointment, appointment)
		}
	}
	var data []SendEmails
	for _, appointment := range upcoming_appointment {
		doctor, err := server.Services.DoctorService.Find(appointment.Doctorid)
		if err != nil {
			if err == sql.ErrNoRows {
				server.Redis.Del(server.Context, strconv.Itoa(appointment.Appointmentid))
			}
		}
		patient, err := server.Services.PatientService.Find(appointment.Patientid)
		if err != nil {
			if err == sql.ErrNoRows {
				server.Redis.Del(server.Context, strconv.Itoa(appointment.Appointmentid))
			}
		}
		type emaildata struct {
			Email          string
			LinkedUsername string
			Date           time.Time
			Username       string
		}
		subject := "Upcoming Appointments!!"
		patientemaildata := server.Mailer.setdata(emaildata{
			Email:          patient.Email,
			Date:           appointment.Appointmentdate,
			LinkedUsername: doctor.Username,
			Username:       patient.Username,
		}, subject, "reminder.template.html", patient.Email)
		doctoremaildata := server.Mailer.setdata(emaildata{
			Email:          doctor.Email,
			LinkedUsername: patient.Username,
			Date:           appointment.Appointmentdate,
			Username:       doctor.Username,
		}, subject, "reminder.template.html", doctor.Email)
		data = append(data, patientemaildata, doctoremaildata)
		server.Redis.Del(server.Context, strconv.Itoa(appointment.Appointmentid))
	}
	go func() {
		for _, notifications := range data {
			notify := notifications
			server.Worker.Task <- &notify
		}
	}()
	server.WaitGroup.Add(server.Worker.Nworker)
	for i := 0; i < server.Worker.Nworker; i++ {
		go func() {
			defer server.Done()
			server.Worker.Workqueue()
		}()
	}
}

func (server *Server) StaffCreateRecord(w http.ResponseWriter, r *http.Request) {
	// BUG: A doctor who doesn't have an appointment with the said patient can create a record!!!!!
	// TODO: Might not Ideal but a fix woould to loop the appointments and check if there's an appointment with the said subjects
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
	msg = NewForm(r, &register)
	data := struct {
		User   DoctorResp
		Errors Errors
		Csrf   map[string]interface{}
	}{
		User:   staff,
		Errors: msg.Errors,
		Csrf:   msg.Csrf,
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "staff-edit-records.html", data)
		return
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
		server.Templates.Render(w, "404.html", nil)
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
	if user.Id != data.Doctorid {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}
	pdata := struct {
		User    DoctorResp
		Errors  Errors
		Records models.Patientrecords
		Csrf    map[string]interface{}
	}{
		Errors:  Errmap,
		Records: data,
		User:    user,
		Csrf:    msg.Csrf,
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
	msg = NewForm(r, &register)
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
		Csrf   map[string]interface{}
	}{
		User:   user,
		Csrf:   msg.Csrf,
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
	http.Redirect(w, r, r.URL.String(), 301)
}

func (server *Server) UploadAvatar(file multipart.File, userid, typeuser, filename string) (string, error) {
	dir := "upload/" + typeuser + "/" + userid + "/" + filepath.Base(filename)
	err := os.MkdirAll(filepath.Dir(dir), 0750)
	if err != nil && !os.IsExist(err) {
		return "", err
	}
	f, err := os.Create(dir)
	defer f.Close()
	_, err = io.Copy(f, file)
	if err != nil {
		return "", err
	}
	fullpath := dir
	return fullpath, nil
}
