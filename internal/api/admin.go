package api

import (
	"database/sql"
	"strconv"

	// "log"
	"net/http"
	// "net/url"
	//
	// "github.com/DataDog/datadog-agent/pkg/clusteragent/admission/mutate"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/patienttracker/internal/models"
	// "github.com/patienttracker/internal/services"
)

const PageCount = 20

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
		server.Templates.Render(w, "admin-login.html", msg)
		return
	}
	msg = Form{
		Data: &login,
	}
	if ok := msg.Validate(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-login.html", msg)
		return
	}
	user, err := server.Services.RbacService.UsersService.FindbyEmail(login.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusBadRequest)
			msg.Errors["Login"] = "No such user"
			server.Templates.Render(w, "admin-login.html", msg)
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

type Pagination struct {
	Page      int
	PrevPage  int
	NextPage  int
	HasPrev   bool
	HasNext   bool
	NextPages []int
	PrevPages []int
}

func (server *Server) Adminrecord(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}
	count, err := server.Services.PatientRecordService.Count()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	params := mux.Vars(r)
	id := params["pageid"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	offset := idparam * PageCount
	paging := Pagination{}
	nextpage := func(id int) int {
		if idparam*PageCount >= count {
			paging.HasNext = false
			return id - 1
		}
		paging.HasNext = true
		return id + 1
	}

	paging.Page = idparam
	paging.NextPage = nextpage(idparam)
	val := func(id int) int {
		if id <= 0 {
			paging.HasPrev = false
			return 0
		}
		paging.HasPrev = true
		return id - 1
	}
	paging.PrevPage = val(idparam)
	records, err := server.Services.PatientRecordService.FindAll(
		models.ListPatientRecords{
			Limit:  PageCount,
			Offset: offset,
		})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}

	data := struct {
		User       UserResp
		Records    []models.Patientrecords
		Pagination Pagination
	}{
		User:       admin,
		Records:    records,
		Pagination: paging,
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "admin-records.html", data)
	return
}

func (server *Server) Adminappointments(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	// w.WriteHeader(http.StatusOK)
	params := mux.Vars(r)
	id := params["pageid"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	count, err := server.Services.AppointmentService.Count()

	offset := idparam * PageCount
	paging := Pagination{}
	nextpage := func(id int) int {
		if idparam*PageCount >= count {
			paging.HasNext = false
			return id - 1
		}
		paging.HasNext = true
		return id + 1
	}

	paging.Page = idparam
	paging.NextPage = nextpage(idparam)
	val := func(id int) int {
		if id <= 0 {
			paging.HasPrev = false
			return 0
		}
		paging.HasPrev = true
		return id - 1
	}
	paging.PrevPage = val(idparam)
	appointment, err := server.Services.AppointmentService.FindAll(models.ListAppointments{
		Limit:  PageCount,
		Offset: offset,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	data := struct {
		User       UserResp
		Apntmt     []models.Appointment
		Pagination Pagination
	}{
		User:       admin,
		Apntmt:     appointment,
		Pagination: paging,
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "admin-appointment.html", data)
	return

}

func (server *Server) Adminuser(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	users, err := server.Services.RbacService.UsersService.FindAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	data := struct {
		User  UserResp
		Users []models.Users
		// Pagination Pagination
	}{
		User:  admin,
		Users: users,
		// Pagination: paging,
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "admin-user.html", data)
	return

}

func (server *Server) Adminroles(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	roles, err := server.Services.RbacService.RolesService.FindAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	data := struct {
		User  UserResp
		Roles []models.Roles
		// Pagination Pagination
	}{
		User:  admin,
		Roles: roles,
		// Pagination: paging,
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "admin-roles.html", data)
	return

}

func (server *Server) Adminpatient(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	params := mux.Vars(r)
	id := params["pageid"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	count, err := server.Services.PatientService.Count()

	offset := idparam * PageCount
	paging := Pagination{}
	nextpage := func(id int) int {
		if idparam*PageCount >= count {
			paging.HasNext = false
			return id - 1
		}
		paging.HasNext = true
		return id + 1
	}

	paging.Page = idparam
	paging.NextPage = nextpage(idparam)
	val := func(id int) int {
		if id <= 0 {
			paging.HasPrev = false
			return 0
		}
		paging.HasPrev = true
		return id - 1
	}
	paging.PrevPage = val(idparam)
	patient, err := server.Services.PatientService.FindAll(models.ListPatients{
		Limit:  PageCount,
		Offset: offset,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	data := struct {
		User       UserResp
		Patient    []models.Patient
		Pagination Pagination
	}{
		User:       admin,
		Patient:    patient,
		Pagination: paging,
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "admin-patient.html", data)
	return

}

func (server *Server) Adminphysician(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	// w.WriteHeader(http.StatusOK)
	params := mux.Vars(r)
	id := params["pageid"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	count, err := server.Services.DoctorService.Count()
	offset := idparam * PageCount
	paging := Pagination{}
	nextpage := func(id int) int {
		if idparam*PageCount >= count {
			paging.HasNext = false
			return id - 1
		}
		paging.HasNext = true
		return id + 1
	}

	paging.Page = idparam
	paging.NextPage = nextpage(idparam)
	val := func(id int) int {
		if id <= 0 {
			paging.HasPrev = false
			return 0
		}
		paging.HasPrev = true
		return id - 1
	}
	paging.PrevPage = val(idparam)
	doctors, err := server.Services.DoctorService.FindAll(models.ListDoctors{
		Limit:  PageCount,
		Offset: offset,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	data := struct {
		User       UserResp
		Doctors    []models.Physician
		Pagination Pagination
	}{
		User:       admin,
		Doctors:    doctors,
		Pagination: paging,
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "admin-physician.html", data)
	return

}

func (server *Server) Adminschedule(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	// w.WriteHeader(http.StatusOK)
	params := mux.Vars(r)
	id := params["pageid"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	count, err := server.Services.ScheduleService.Count()

	offset := idparam * PageCount
	paging := Pagination{}
	nextpage := func(id int) int {
		if idparam*PageCount >= count {
			paging.HasNext = false
			return id - 1
		}
		paging.HasNext = true
		return id + 1
	}

	paging.Page = idparam
	paging.NextPage = nextpage(idparam)
	val := func(id int) int {
		if id <= 0 {
			paging.HasPrev = false
			return 0
		}
		paging.HasPrev = true
		return id - 1
	}
	paging.PrevPage = val(idparam)
	schedules, err := server.Services.ScheduleService.FindAll(models.ListSchedules{
		Limit:  PageCount,
		Offset: offset,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	data := struct {
		User       UserResp
		Schedule   []models.Schedule
		Pagination Pagination
	}{
		User:       admin,
		Schedule:   schedules,
		Pagination: paging,
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "admin-schedule.html", data)
	return

}

func (server *Server) Admindepartment(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	// w.WriteHeader(http.StatusOK)
	params := mux.Vars(r)
	id := params["pageid"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	count, err := server.Services.DepartmentService.Count()

	offset := idparam * PageCount
	paging := Pagination{}
	nextpage := func(id int) int {
		if idparam*PageCount >= count {
			paging.HasNext = false
			return id - 1
		}
		paging.HasNext = true
		return id + 1
	}

	paging.Page = idparam
	paging.NextPage = nextpage(idparam)
	val := func(id int) int {
		if id <= 0 {
			paging.HasPrev = false
			return 0
		}
		paging.HasPrev = true
		return id - 1
	}
	paging.PrevPage = val(idparam)
	department, err := server.Services.DepartmentService.FindAll(models.ListDepartment{
		Limit:  PageCount,
		Offset: offset,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	data := struct {
		User       UserResp
		Department []models.Department
		Pagination Pagination
	}{
		User:       admin,
		Department: department,
		Pagination: paging,
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "admin-department.html", data)
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
