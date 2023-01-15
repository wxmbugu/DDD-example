package api

import (
	"database/sql"
	// "log"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/patienttracker/internal/models"
	// "github.com/patienttracker/internal/services"
)

type UserResp struct {
	Id            int
	Email         string
	Authenticated bool
}

func (server *Server) AdminLogin(w http.ResponseWriter, r *http.Request) {
	var msg Form
	session, err := server.Store.Get(r, "admin")
	if err = session.Save(r, w); err != nil {
		http.Redirect(w, r, "/500", 300)

	}
	login := Login{
		Email:    r.PostFormValue("email"),
		Password: r.PostFormValue("password"),
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "login.html", nil)
		return
	}
	msg = Form{
		Data: &login,
	}
	if ok := msg.Validate(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "login.html", msg)
		return
	}
	user, err := server.Services.RbacService.UsersService.FindbyEmail(login.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusBadRequest)
			msg.Errors["Login"] = "No such user"
			server.Templates.Render(w, "login.html", msg)
			return
		}
		http.Redirect(w, r, "/500", 300)
	}
	// if err = services.CheckPassword(user.Password, login.Password); err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	msg.Errors["Login"] = "No such user"
	// 	server.Templates.Render(w, "login.html", msg)
	// 	return
	// }
	if user.Password == login.Password {
		user := models.Users{
			Id:    user.Id,
			Email: user.Email,
		}
		admin := UserResponse(user)
		gobRegister(admin)
		session.Values["admin"] = admin
		if err = session.Save(r, w); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Redirect(w, r, "/500", 300)

		}
		http.Redirect(w, r, "/admin/home", 300)

	}
}

func UserResponse(user models.Users) UserResp {
	return UserResp{
		Email:         user.Email,
		Id:            user.Id,
		Authenticated: true,
	}
}
func (server *Server) Adminhome(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	user := getAdmin(session)
	if !user.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	w.WriteHeader(http.StatusOK)

	appointment, err := server.Services.AppointmentService.FindAll(models.ListAppointments{
		Limit:  10000,
		Offset: 1,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	records, err := server.Services.PatientRecordService.FindAll(
		models.ListPatientRecords{
			Limit:  10000,
			Offset: 1,
		})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	data := struct {
		User    UserResp
		Apntmt  []models.Appointment
		Records []models.Patientrecords
	}{
		User:    user,
		Apntmt:  appointment,
		Records: records,
	}
	// log.Println(data.Records)
	server.Templates.Render(w, "admin-home.html", data)
	return

}

func getAdmin(s *sessions.Session) UserResp {
	val := s.Values["admin"]
	var user = UserResp{}
	user, ok := val.(UserResp)
	if !ok {
		return UserResp{Authenticated: false}
	}
	return user
}
