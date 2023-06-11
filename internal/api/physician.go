package api

import (
	"database/sql"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
	session, _ := server.Store.Get(r, "staff")
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
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
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
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	http.Redirect(w, r, "/staff/home", http.StatusSeeOther)
}
func (server *Server) Staffcreateschedule(w http.ResponseWriter, r *http.Request) {
	session, _ := server.Store.Get(r, "staff")
	user := getStaff(session)
	if !user.Authenticated {
		http.Redirect(w, r, "/staff/login", http.StatusMovedPermanently)
	}
	var msg Form
	var actvie bool
	register := Schedule{
		Doctorid:  r.PostFormValue("Doctorid"),
		Starttime: r.PostFormValue("Starttime"),
		Endtime:   r.PostFormValue("Endtime"),
	}
	msg = NewForm(r, &register)
	data := struct {
		User    DoctorResp
		Active  []string
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
	actvie = checkboxvalue(r.PostFormValue("Active"))
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
	w.WriteHeader(http.StatusCreated)
	data.Success = "schedule created successfuly"
	server.Templates.Render(w, "staff-edit-schedule.html", data)
}
func (server *Server) StaffLogout(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "staff")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	session.Values["staff"] = DoctorResp{}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	http.Redirect(w, r, "/staff/home", http.StatusMovedPermanently)
}

func (server *Server) staffviewrecord(w http.ResponseWriter, r *http.Request) {
	errmap := make(map[string]string)
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	data, err := server.Services.PatientRecordService.Find(idparam)
	if err != nil {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	session, err := server.Store.Get(r, "staff")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user := getStaff(session)
	if !user.Authenticated {
		http.Redirect(w, r, "/staff/login", http.StatusMovedPermanently)
	}
	if user.Id != data.Doctorid {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	pdata := struct {
		User    DoctorResp
		Errors  Errors
		Records models.Patientrecords
	}{
		Errors:  errmap,
		Records: data,
		User:    user,
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "staff-update-record.html", pdata)
}

func (server *Server) Staffupdateschedule(w http.ResponseWriter, r *http.Request) {
	var msg Form
	Errmap := make(map[string]string)
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	data, err := server.Services.ScheduleService.Find(idparam)
	if err != nil {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	session, err := server.Store.Get(r, "staff")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user := getStaff(session)
	if !user.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
	}
	register := Schedule{
		Doctorid:  r.PostFormValue("Doctorid"),
		Starttime: r.PostFormValue("Starttime"),
		Endtime:   r.PostFormValue("Endtime"),
	}
	msg = NewForm(r, &register)
	pdata := struct {
		User     DoctorResp
		Errors   Errors
		Csrf     map[string]interface{}
		Schedule models.Schedule
		Success  string
	}{
		Errors:   Errmap,
		Schedule: data,
		Csrf:     msg.Csrf,
		User:     user,
	}
	var active bool
	if r.Method == http.MethodGet {
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
	active = checkboxvalue(r.PostFormValue("Active"))
	schedule := models.Schedule{
		Scheduleid: data.Scheduleid,
		Doctorid:   doctorid,
		Starttime:  register.Starttime,
		Endtime:    register.Endtime,
		Active:     active,
	}
	if _, err := server.Services.UpdateSchedule(schedule); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		pdata.Errors = Errmap
		server.Templates.Render(w, "staff-update-schedule.html", pdata)
		return
	}
	w.WriteHeader(http.StatusOK)
	pdata.Success = "schedule updated"
	pdata.Schedule = schedule
	server.Templates.Render(w, "staff-update-schedule.html", pdata)
}
func (server *Server) Staffschedule(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "staff")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user := getStaff(session)
	if !user.Authenticated {
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	}

	schedules, err := server.Services.ScheduleService.FindbyDoctor(user.Id)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
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
}

func (server *Server) Staffdeleteschedule(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "staff")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user := getStaff(session)
	if !user.Authenticated {
		http.Redirect(w, r, "/staff/login", http.StatusMovedPermanently)
	}
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	if err := server.Services.ScheduleService.Delete(idparam); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "staff-schedule.html", nil)
		return
	}
	http.Redirect(w, r, "/staff/schedules", http.StatusMovedPermanently)
}
func (server *Server) Staffprofile(w http.ResponseWriter, r *http.Request) {
	Errmap := make(map[string]string)
	session, err := server.Store.Get(r, "staff")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user := getStaff(session)
	if !user.Authenticated {
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	}
	doc, err := server.Services.DoctorService.Find(user.Id)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
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
		User    DoctorResp
		Doctor  models.Physician
		Errors  Errors
		Csrf    map[string]interface{}
		Success string
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
	w.WriteHeader(http.StatusOK)
	data.Doctor = doctor
	data.Success = "account updated successfully"
	server.Templates.Render(w, "doctor-profile.html", data)
}
func (server *Server) Staffhome(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "staff")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user := getStaff(session)
	if !user.Authenticated {
		http.Redirect(w, r, "/staff/login", http.StatusMovedPermanently)
	}
	appointment, err := server.Services.AppointmentService.FindAllByDoctor(user.Id)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	records, err := server.Services.PatientRecordService.FindAllByDoctor(user.Id)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
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
}

func (server *Server) Staffappointments(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "staff")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user := getStaff(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/staff/login", http.StatusSeeOther)
		return
	}
	appointment, err := server.Services.AppointmentService.FindAllByDoctor(user.Id)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	data := struct {
		User   DoctorResp
		Apntmt []models.Appointment
	}{
		User:   user,
		Apntmt: appointment,
	}

	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "staff-appointments.html", data)
}
func (server *Server) Staffrecord(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "staff")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user := getStaff(session)
	if !user.Authenticated {
		http.Redirect(w, r, "/staff/login", http.StatusMovedPermanently)
	}

	records, err := server.Services.PatientRecordService.FindAllByDoctor(user.Id)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	data := struct {
		User    DoctorResp
		Records []models.Patientrecords
	}{
		User:    user,
		Records: records,
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "staff-records.html", data)
}

func (server *Server) StaffUpdateAppointment(w http.ResponseWriter, r *http.Request) {
	var msg Form
	Errmap := make(map[string]string)
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	data, err := server.Services.AppointmentService.Find(idparam)
	if err != nil {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	session, err := server.Store.Get(r, "staff")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user := getStaff(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/staff/login", http.StatusSeeOther)
		return
	}
	if user.Id != data.Doctorid {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	register := Appointment{
		Doctorid:        r.PostFormValue("Doctorid"),
		Patientid:       r.PostFormValue("Patientid"),
		AppointmentDate: r.PostFormValue("Appointmentdate"),
		Duration:        r.PostFormValue("Duration"),
	}
	msg = NewForm(r, &register)
	pdata := struct {
		User        DoctorResp
		Errors      Errors
		Appointment models.Appointment
		Csrf        map[string]interface{}
		Success     string
	}{
		Errors:      Errmap,
		Appointment: data,
		User:        user,
		Csrf:        msg.Csrf,
	}
	var approval bool
	if r.Method == http.MethodGet {
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
	doctorid, _ := strconv.Atoi(r.PostFormValue("Doctorid"))
	patientid, _ := strconv.Atoi(r.PostFormValue("Patientid"))
	date, err := time.Parse("2006-01-02T15:04", r.PostFormValue("Appointmentdate"))
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	var outbound bool
	approval = checkboxvalue(r.PostFormValue("Approval"))
	outbound = checkboxvalue(r.PostFormValue("Outbound"))
	apntmt := models.Appointment{
		Appointmentid:   data.Appointmentid,
		Doctorid:        doctorid,
		Patientid:       patientid,
		Appointmentdate: date,
		Duration:        register.Duration,
		Outbound:        outbound,
		Approval:        approval,
	}
	appointment, err := server.Services.UpdateappointmentbyPatient(apntmt)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		pdata.Errors = Errmap
		server.Templates.Render(w, "staff-update-appointment.html", pdata)
		return
	}
	if appointment.Approval {
		if err := server.Redis.Set(server.Context, strconv.Itoa(appointment.Appointmentid), appointment.Appointmentdate, 0).Err(); err != nil {
			server.Log.Error(err)
		}
	}
	w.WriteHeader(http.StatusOK)
	pdata.Appointment = apntmt
	pdata.Success = "appointment updated successfully"
	server.Templates.Render(w, "staff-update-appointment.html", pdata)
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
		if int(time.Until(appointment.Appointmentdate).Hours()) <= 24 {
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
		go func(i int) {
			defer server.Done()
			server.Worker.Workqueue()
		}(i)
	}
}

func (server *Server) UploadAvatar(file multipart.File, userid, typeuser, filename string) (string, error) {
	dir := "upload/" + typeuser + "/" + userid + "/" + filepath.Base(filename)
	err := os.MkdirAll(filepath.Dir(dir), 0750)
	if err != nil && !os.IsExist(err) {
		return "", err
	}
	f, err := os.Create(dir)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = io.Copy(f, file)
	if err != nil {
		return "", err
	}
	fullpath := dir
	return fullpath, nil
}

func (server *Server) doctor_reset_password(w http.ResponseWriter, r *http.Request) {
	Errmap := make(map[string]string)
	id := r.URL.Query().Get("id")
	if !strings.Contains(id, "doctor") {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	value, err := server.Redis.Get(server.Context, id).Result()
	if err != nil {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	doctor, err := server.Services.DoctorService.FindbyEmail(value)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Redirect(w, r, "/404", http.StatusMovedPermanently)
		}
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	register := ResetPassword{
		Email:           r.PostFormValue("Email"),
		Password:        r.PostFormValue("Password"),
		ConfirmPassword: r.PostFormValue("ConfirmPassword"),
	}
	msg := NewForm(r, &register)
	data := struct {
		Doctor     models.Physician
		Errors     Errors
		Csrf       map[string]interface{}
		Bloodgroup []string
		Success    string
	}{
		Errors:     Errmap,
		Doctor:     doctor,
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
	doctor.Password_changed_at = time.Now()
	doctor.Hashed_password, err = services.HashPassword(register.Password)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	if _, err := server.Services.DoctorService.Update(doctor); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		data.Errors = Errmap
		server.Templates.Render(w, "password_reset.html", data)
		return
	}
	w.WriteHeader(http.StatusOK)
	data.Success = "password reset successfully"
	server.Templates.Render(w, "password_reset.html", data)
}
