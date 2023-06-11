package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	// "strings"
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
			http.Redirect(w, r, "/404", http.StatusMovedPermanently)
		}
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
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
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	http.Redirect(w, r, "/nurse/home", http.StatusMovedPermanently)
}

func (server *Server) NurseCreateRecord(w http.ResponseWriter, r *http.Request) {
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
	patientid, _ := strconv.Atoi(r.PostFormValue("Patientid"))
	doctorid, _ := strconv.Atoi(r.PostFormValue("Doctorid"))
	register := Records{
		Height:      r.PostFormValue("Height"),
		Bp:          r.PostFormValue("Bp"),
		Temperature: r.PostFormValue("Temperature"),
		Weight:      r.PostFormValue("Weight"),
		Patientid:   r.PostFormValue("Patientid"),
		Doctorid:    r.PostFormValue("Doctorid"),
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
		Doctorid:    doctorid,
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
	w.WriteHeader(http.StatusCreated)
	data.Success = "record created successfuly"
	server.Templates.Render(w, "nurse-edit-record.html", data)
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
		http.Redirect(w, r, "/staff/login", http.StatusMovedPermanently)
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
	for i := 0; i < server.Worker.Nworker; i++ {
		go func(i int) {
			server.Worker.Task <- &mailer
			defer server.Done()
			server.Worker.Workqueue()
		}(i)
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
}
