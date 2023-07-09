package api

import (
	"database/sql"
	"fmt"
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
	session, _ := server.Store.Get(r, "nurse")
	login := Login{
		Email:    r.PostFormValue("email"),
		Password: r.PostFormValue("password"),
	}
	msg = NewForm(r, &login)
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "nurse-login.html", msg)
		return
	}
	if ok := msg.Validate(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "nurse-login.html", msg)
		return
	}
	nurse, err := server.Services.NurseService.FindbyEmail(login.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusBadRequest)
			msg.Errors["Login"] = "No such user"
			server.Templates.Render(w, "nurse-login.html", msg)
			return
		}
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	if err = services.CheckPassword(nurse.Hashed_password, login.Password); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Login"] = "No such user"
		server.Templates.Render(w, "nurse-login.html", msg)
		return
	}
	user := NurseResponse(nurse)
	gobRegister(user)
	session.Values["nurse"] = user
	if err = session.Save(r, w); err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	http.Redirect(w, r, "/nurse/home", http.StatusMovedPermanently)
}
func (server *Server) NurseLogout(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "nurse")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	session.Values["nurse"] = NurseResp{}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	http.Redirect(w, r, "/nurse/login", http.StatusMovedPermanently)
}
func (server *Server) Nurserecord(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "nurse")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user := getNurse(session)
	if !user.Authenticated {
		http.Redirect(w, r, "/staff/login", http.StatusMovedPermanently)
	}
	w.WriteHeader(http.StatusOK)

	records, err := server.Services.PatientRecordService.FindAllByNurse(user.Id)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	data := struct {
		User    NurseResp
		Records []models.Patientrecords
	}{
		User:    user,
		Records: records,
	}
	server.Templates.Render(w, "nurse-records.html", data)
}
func (server *Server) NurseViewRecord(w http.ResponseWriter, r *http.Request) {
	errmap := make(map[string]string)
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	data, err := server.Services.PatientRecordService.Find(idparam)
	if err != nil {
		server.Templates.Render(w, "404.html", nil)
	}
	session, err := server.Store.Get(r, "nurse")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user := getNurse(session)
	if !user.Authenticated {
		http.Redirect(w, r, "/nurse/login", http.StatusMovedPermanently)
	}
	if user.Id != data.Nurseid {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}
	pdata := struct {
		User    NurseResp
		Errors  Errors
		Records models.Patientrecords
	}{
		Errors:  errmap,
		Records: data,
		User:    user,
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "nurse-view-record.html", pdata)
}

func (server *Server) resetpassword(w http.ResponseWriter, r *http.Request) {
	reset := Reset{
		Email: r.FormValue("Email"),
	}
	var form = NewForm(r, &reset)
	dt := struct {
		Success string
		Csrf    map[string]interface{}
		Errors  Errors
	}{
		Csrf: form.Csrf,
	}
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "reset.html", dt)
		return
	}
	if ok := form.Validate(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		dt.Errors = form.Errors
		server.Templates.Render(w, "reset.html", dt)
		return
	}
	key, account_type := key_reset_pass(r.URL.String())
	value := reset.Email
	if err := server.Redis.Set(server.Context, key, value, 0).Err(); err != nil {
		server.Log.Error(err)
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	data := struct {
		URL   string
		Email string
	}{
		URL:   fmt.Sprintf(`http://localhost:9000/%s/passwordreset?id=%s`, account_type, key),
		Email: reset.Email,
	}
	mailer := server.Mailer.setdata(data, "Reset Password!!", "reset_password.account.html", data.Email)
	server.WaitGroup.Add(server.Worker.Nworker)
	go func() {
		server.Worker.Task <- &mailer
	}()
	server.WaitGroup.Add(server.Worker.Nworker)
	for i := 0; i < server.Worker.Nworker; i++ {
		go func() {
			defer server.Done()
			server.Worker.Workqueue()
		}()
	}
	w.WriteHeader(http.StatusCreated)
	dt.Success = "email sent successfuly to reset your account"
	server.Templates.Render(w, "reset.html", dt)
}
func key_reset_pass(url string) (string, string) {
	if strings.Contains(url, "nurse") {
		return "nurse" + utils.RandString(40), "nurse"
	} else if strings.Contains(url, "doctor") {
		return "doctor" + utils.RandString(40), "doctor"
	} else if strings.Contains(url, "admin") {
		return "admin" + utils.RandString(40), "admin"
	}
	return "patient" + utils.RandString(40), "patient"
}

func (server *Server) nurse_reset_password(w http.ResponseWriter, r *http.Request) {
	Errmap := make(map[string]string)
	id := r.URL.Query().Get("id")
	if !strings.Contains(id, "nurse") {
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
	nurse, err := server.Services.NurseService.FindbyEmail(value)
	if err != nil {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	register := ResetPassword{
		Email:           r.PostFormValue("Email"),
		Password:        r.PostFormValue("Password"),
		ConfirmPassword: r.PostFormValue("ConfirmPassword"),
	}
	msg := NewForm(r, &register)
	data := struct {
		Nurse      models.Nurse
		Errors     Errors
		Csrf       map[string]interface{}
		Bloodgroup []string
		Success    string
	}{
		Errors:     Errmap,
		Nurse:      nurse,
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
	nurse.Password_changed_at = time.Now()
	nurse.Hashed_password, err = services.HashPassword(register.Password)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	if _, err := server.Services.NurseService.Update(nurse); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		data.Errors = Errmap
		server.Templates.Render(w, "password_reset.html", data)
		return
	}
	w.WriteHeader(http.StatusOK)
	data.Success = "account updated successfuly"
	server.Templates.Render(w, "password_reset.html", data)
	server.Redis.Del(server.Context, id)
}

func (server *Server) Nursetickets(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "nurse")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user := getNurse(session)
	if !user.Authenticated {
		http.Redirect(w, r, "/nurse/login", http.StatusMovedPermanently)
	}
	var tickets = server.getnursetickets(user.Id)
	data := struct {
		User    NurseResp
		Tickets []Ticket
	}{
		User:    user,
		Tickets: tickets,
	}
	server.Templates.Render(w, "nurse-tickets.html", data)
}

func (server *Server) getnursetickets(nurseid int) []Ticket {
	var ticketids []string
	var t = Ticket{}
	var tickets []Ticket
	iter := server.Redis.Scan(server.Context, 0, "*", 0).Iterator()
	for iter.Next(server.Context) {
		if strings.Contains(iter.Val(), "ticket") {
			ticketids = append(ticketids, iter.Val())
		}
	}
	for _, ids := range ticketids {
		value, err := server.Redis.Get(server.Context, ids).Result()
		if err != nil {
			return tickets
		}
		t.UnMarshalBinary([]byte(value))
		if t.Nurseid == nurseid {
			tickets = append(tickets, t)
		}
	}
	return tickets
}

func (server *Server) NurseCreateRecord(w http.ResponseWriter, r *http.Request) {
	var t Ticket
	session, err := server.Store.Get(r, "nurse")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	nurse := getNurse(session)
	if !nurse.Authenticated {
		http.Redirect(w, r, "/staff/login", http.StatusMovedPermanently)
	}
	var msg Form
	height, _ := strconv.Atoi(r.PostFormValue("Height"))
	bp, _ := strconv.Atoi(r.PostFormValue("Bp"))
	temp, _ := strconv.Atoi(r.PostFormValue("Temperature"))
	heartrate, _ := strconv.Atoi(r.PostFormValue("HeartRate"))
	ticketid := mux.Vars(r)
	ticket, err := server.Redis.Get(server.Context, ticketid["ticket"]).Result()
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	t.UnMarshalBinary([]byte(ticket))
	patient, err := server.Services.PatientService.FindbyEmail(t.Patientemail)
	if err != nil {
		if err == sql.ErrNoRows {
			if err != nil {
				http.Redirect(w, r, "/404", http.StatusMovedPermanently)
			}
		}
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	register := Records{
		Height:      r.PostFormValue("Height"),
		Bp:          r.PostFormValue("Bp"),
		Temperature: r.PostFormValue("Temperature"),
		Weight:      r.PostFormValue("Weight"),
		Patientid:   strconv.Itoa(patient.Patientid),
		Doctorid:    strconv.Itoa(t.Doctorid),
		HeartRate:   r.PostFormValue("HeartRate"),
	}
	msg = NewForm(r, &register)
	data := struct {
		User    NurseResp
		Errors  Errors
		Csrf    map[string]interface{}
		Success string
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
		Doctorid:    t.Doctorid,
		Patienid:    patient.Patientid,
		Nurseid:     nurse.Id,
		Height:      height,
		Bp:          bp,
		Temperature: temp,
		HeartRate:   heartrate,
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
	w.WriteHeader(http.StatusCreated)
	data.Success = "record created successfuly"
	server.Templates.Render(w, "nurse-edit-record.html", data)
}

func (server *Server) Nurseprofile(w http.ResponseWriter, r *http.Request) {
	Errmap := make(map[string]string)
	session, err := server.Store.Get(r, "nurse")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user := getNurse(session)
	if !user.Authenticated {
		http.Redirect(w, r, "/nurse/login", http.StatusMovedPermanently)
	}
	nusrse, err := server.Services.NurseService.Find(user.Id)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}

	register := NurseRegister{
		Email:           r.PostFormValue("Email"),
		Password:        r.PostFormValue("Password"),
		ConfirmPassword: r.PostFormValue("ConfirmPassword"),
		Username:        r.PostFormValue("Username"),
		Fullname:        r.PostFormValue("Fullname"),
	}
	msg := NewForm(r, &register)
	data := struct {
		User    NurseResp
		Nurse   models.Nurse
		Errors  Errors
		Csrf    map[string]interface{}
		Success string
	}{
		User:  user,
		Nurse: nusrse,
		Csrf:  msg.Csrf,
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "nurse-profile.html", data)
		return
	}
	if ok := msg.Validate(); !ok {
		data.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "nurse-profile.html", data)
		return
	}

	hashed_password, _ := services.HashPassword(register.Password)
	nurse := models.Nurse{
		Id:                  user.Id,
		Username:            register.Username,
		Full_name:           register.Fullname,
		Email:               register.Email,
		Hashed_password:     hashed_password,
		Password_changed_at: time.Now(),
	}
	if _, err := server.Services.NurseService.Update(nurse); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		data.Errors = Errmap
		server.Templates.Render(w, "nurse-profile.html", data)
		return
	}
	w.WriteHeader(http.StatusOK)
	data.Nurse = nurse
	data.Success = "account updated successfully"
	server.Templates.Render(w, "nurse-profile.html", data)
}
