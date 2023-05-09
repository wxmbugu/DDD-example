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
)

const PageCount = 20

type UserResp struct {
	Id            int
	Email         string
	Authenticated bool
	Permission    []string
}

func (server *Server) AdminLogin(w http.ResponseWriter, r *http.Request) {
	var msg Form
	login := Login{
		Email:    r.PostFormValue("email"),
		Password: r.PostFormValue("password"),
	}
	msg = NewForm(r, &login)
	session, err := server.Store.Get(r, "admin")
	if err = session.Save(r, w); err != nil {
		http.Redirect(w, r, "/500", 300)

	}

	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "admin-login.html", msg)
		return
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
	if err = services.CheckPassword(user.Password, login.Password); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Login"] = "No such user"
		server.Templates.Render(w, "login.html", msg)
		return
	}
	permission, err := server.Services.RbacService.PermissionsService.FindbyRoleId(user.Roleid)
	user = models.Users{
		Id:    user.Id,
		Email: user.Email,
	}
	var perm []string
	for _, v := range permission {
		perm = append(perm, v.Permission)
	}
	admin := UserResponse(user, perm)
	gobRegister(admin)
	session.Values["admin"] = admin
	if err = session.Save(r, w); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		http.Redirect(w, r, "/500", 300)
	}
	http.Redirect(w, r, "/admin/home", 300)
}
func (server *Server) AdminLogout(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	session.Values["admin"] = UserResp{}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	http.Redirect(w, r, "/admin/home", 300)
}

func UserResponse(user models.Users, permmission []string) UserResp {
	return UserResp{
		Email:         user.Email,
		Id:            user.Id,
		Authenticated: true,
		Permission:    permmission,
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		data := strings.Split(v, ":")
		if data[0] == "record" {
			acceptedperm = append(acceptedperm, v)
		}
		if v == "admin" || v == "editor" || v == "viewwer" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		data := strings.Split(v, ":")
		if data[0] == "appointment" {
			acceptedperm = append(acceptedperm, v)
		}
		if v == "admin" || v == "editor" || v == "viewwer" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
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

func (server *Server) AdminNurses(w http.ResponseWriter, r *http.Request) {
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" || v == "editor" || v == "viewer" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}
	params := mux.Vars(r)
	id := params["pageid"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	count, err := server.Services.NurseService.Count()
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
	nurse, err := server.Services.NurseService.FindAll(models.ListNurses{
		Limit:  PageCount,
		Offset: offset,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	data := struct {
		User       UserResp
		Nurse      []models.Nurse
		Pagination Pagination
	}{
		User:       admin,
		Nurse:      nurse,
		Pagination: paging,
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "admin-nurse.html", data)
	return
}

func (server *Server) Admincreateuser(w http.ResponseWriter, r *http.Request) {
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}
	var msg Form
	register := AdminstrativeUser{
		Email:           r.PostFormValue("Email"),
		Rolename:        r.PostFormValue("Rolename"),
		Password:        r.PostFormValue("Password"),
		ConfirmPassword: r.PostFormValue("ConfirmPassword"),
	}
	msg = NewForm(r, &register)
	data := struct {
		User   UserResp
		Errors Errors
		Csrf   map[string]interface{}
	}{
		User:   admin,
		Errors: msg.Errors,
		Csrf:   msg.Csrf,
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "admin-edit-user.html", data)
		return
	}
	if ok := msg.Validate(); !ok {
		data.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-edit-user.html", data)
		return
	}
	role, err := server.Services.RbacService.RolesService.FindbyRole(register.Rolename)
	if err != nil {
		if err == sql.ErrNoRows {
			msg.Errors["NonExistence"] = "No such role"
			data.Errors = msg.Errors
			w.WriteHeader(http.StatusBadRequest)
			server.Templates.Render(w, "admin-edit-user.html", data)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	password, err := services.HashPassword(register.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	if _, err := server.Services.RbacService.UsersService.Create(models.Users{
		Email:    register.Email,
		Password: password,
		Roleid:   role.Roleid,
	}); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Exists"] = err.Error()
		data.Errors = msg.Errors
		server.Templates.Render(w, "admin-edit-user.html", data)
		return
	}
	http.Redirect(w, r, "/admin/users", 301)
	return
}

func (server *Server) Adminupdateuser(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}
	var msg Form
	register := AdminstrativeUser{
		Email:           r.PostFormValue("Email"),
		Rolename:        r.PostFormValue("Rolename"),
		Password:        r.PostFormValue("Password"),
		ConfirmPassword: r.PostFormValue("ConfirmPassword"),
	}
	msg = NewForm(r, &register)
	user, err := server.Services.RbacService.UsersService.Find(idparam)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	role, err := server.Services.RbacService.RolesService.Find(user.Roleid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	data := struct {
		User      UserResp
		Errors    Errors
		Csrf      map[string]interface{}
		AdminUser models.Users
		Role      string
	}{
		User:      admin,
		Errors:    msg.Errors,
		Csrf:      msg.Csrf,
		AdminUser: user,
		Role:      role.Role,
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "admin-update-user.html", data)
		return
	}
	if ok := msg.Validate(); !ok {
		data.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-update-user.html", data)
		return
	}
	role, err = server.Services.RbacService.RolesService.FindbyRole(register.Rolename)
	if err != nil {
		if err == sql.ErrNoRows {
			msg.Errors["NonExistence"] = "No such role"
			data.Errors = msg.Errors
			w.WriteHeader(http.StatusBadRequest)
			server.Templates.Render(w, "admin-update-user.html", data)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	password, err := services.HashPassword(register.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	if _, err := server.Services.RbacService.UsersService.Update(models.Users{
		Id:       user.Id,
		Email:    register.Email,
		Password: password,
		Roleid:   role.Roleid,
	}); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Exists"] = err.Error()
		data.Errors = msg.Errors
		server.Templates.Render(w, "admin-update-user.html", data)
		return
	}
	http.Redirect(w, r, "/admin/users", 301)
	return
}

func (server *Server) Admindeleteuser(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}
	if err := server.Services.RbacService.UsersService.Delete(idparam); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	http.Redirect(w, r, "/admin/users", 301)
	return
}
func (server *Server) Admindeleterole(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}
	if err := server.Services.RbacService.RolesService.Delete(idparam); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	http.Redirect(w, r, "/admin/roles", 301)
	return
}
func (server *Server) Admindeletenurse(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" || v == "nurse:admin" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}
	if err := server.Services.NurseService.Delete(idparam); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	http.Redirect(w, r, "/admin/home", 301)
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		data := strings.Split(v, ":")
		if data[0] == "roles" {
			acceptedperm = append(acceptedperm, v)
		}
		if v == "admin" || v == "editor" || v == "viewwer" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
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
	}{
		User:  admin,
		Roles: roles,
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "admin-roles.html", data)
	return

}
func generate_permission() []string {
	var p services.Permissions
	var tablelist = []string{
		"physician",
		"appointment",
		"schedule",
		"patient",
		"department",
		"records",
		"nurse",
	}

	var permission = []string{
		"admin",
		"editor",
		"viewer",
	}
	for _, perm := range permission {
		for _, table := range tablelist {
			value := p.Define(table, services.Str_to_Permission(perm))
			permission = append(permission, value)
		}
	}
	return permission
}
func (server *Server) AdmincreateRoles(w http.ResponseWriter, r *http.Request) {
	available_permissions := generate_permission()
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}
	var msg Form
	register := Role{
		Rolename:   r.PostFormValue("Role"),
		Permission: r.PostFormValue("permission"),
	}
	msg = NewForm(r, &register)
	data := struct {
		User       UserResp
		Errors     Errors
		Permission []string
		Csrf       map[string]interface{}
	}{
		User:       admin,
		Errors:     msg.Errors,
		Permission: available_permissions,
		Csrf:       msg.Csrf,
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "admin-edit-role.html", data)
		return
	}
	if ok := msg.Validate(); !ok {
		data.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-edit-role.html", data)
		return
	}
	role, err := server.Services.RbacService.RolesService.Create(models.Roles{
		Role: r.PostFormValue("Role"),
	},
	)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Exists"] = err.Error()
		data.Errors = msg.Errors
		server.Templates.Render(w, "admin-edit-role.html", data)
		return
	}
	if _, err = server.Services.CreatePermission(models.Permissions{
		Permission: r.PostFormValue("permission"),
		Roleid:     role.Roleid,
	}, admin.Id); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Exists"] = err.Error()
		data.Errors = msg.Errors
		server.Templates.Render(w, "admin-edit-role.html", data)
		return
	}
	http.Redirect(w, r, "/admin/home", 300)
}

func (server *Server) Adminupdateroles(w http.ResponseWriter, r *http.Request) {
	available_permissions := generate_permission()
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" || v == "editor" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}

	var msg Form
	assigned_permissions, err := server.Services.RbacService.PermissionsService.FindbyRoleId(idparam)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	register := UpdateRole{
		Rolename:   r.PostFormValue("Role"),
		Permission: r.Form["permission"],
	}
	msg = NewForm(r, &register)
	role, _ := server.Services.RbacService.RolesService.Find(idparam)
	data := struct {
		User                 UserResp
		Errors               Errors
		Rolename             string
		Permission           []string
		Assigned_Permissions []models.Permissions
		Csrf                 map[string]interface{}
	}{
		User:                 admin,
		Errors:               msg.Errors,
		Rolename:             role.Role,
		Csrf:                 msg.Csrf,
		Permission:           available_permissions,
		Assigned_Permissions: assigned_permissions,
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "admin-update-role.html", data)
		return
	}
	if ok := msg.Validate(); !ok {
		data.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-update-role.html", data)
		return
	}
	_, err = server.Services.RbacService.UsersService.Find(admin.Id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Exists"] = err.Error()
		data.Errors = msg.Errors
		server.Templates.Render(w, "admin-update-role.html", data)
		return

	}
	role, err = server.Services.RbacService.RolesService.Update(
		models.Roles{
			Role:   r.PostFormValue("Role"),
			Roleid: role.Roleid,
		})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Exists"] = err.Error()
		data.Errors = msg.Errors
		server.Templates.Render(w, "admin-update-role.html", data)
		return
	}
	if err := server.Services.UpdateRolePermissions(register.Permission, idparam); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Exists"] = err.Error()
		data.Errors = msg.Errors
		server.Templates.Render(w, "admin-update-role.html", data)
		return
	}
	http.Redirect(w, r, r.URL.String(), 301)
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
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}
	var acceptedperm []string
	for _, v := range admin.Permission {
		data := strings.Split(v, ":")
		if data[0] == "patient" {
			acceptedperm = append(acceptedperm, v)
		}
		if v == "admin" || v == "editor" || v == "viewwer" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
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
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}
	var acceptedperm []string
	for _, v := range admin.Permission {
		data := strings.Split(v, ":")
		if data[0] == "physician" {
			acceptedperm = append(acceptedperm, v)
		}
		if v == "admin" || v == "editor" || v == "viewwer" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
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
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}
	var acceptedperm []string
	for _, v := range admin.Permission {
		data := strings.Split(v, ":")
		if data[0] == "schedule" {
			acceptedperm = append(acceptedperm, v)
		}
		if v == "admin" || v == "editor" || v == "viewwer" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
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
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}
	var acceptedperm []string
	for _, v := range admin.Permission {
		data := strings.Split(v, ":")
		if data[0] == "department" {
			acceptedperm = append(acceptedperm, v)
		}
		if v == "admin" || v == "editor" || v == "viewwer" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
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

func (server *Server) Admincreatepatient(w http.ResponseWriter, r *http.Request) {
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" || v == "editor" || v == "patient:admin" || v == "patient:editor" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}
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
	data := struct {
		User       UserResp
		Errors     Errors
		Csrf       map[string]interface{}
		Bloodgroup []string
	}{
		User:       admin,
		Errors:     msg.Errors,
		Bloodgroup: bloodgroup_array(),
		Csrf:       msg.Csrf,
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "admin-edit-patient.html", data)
		return
	}
	if ok := msg.Validate(); !ok {
		data.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-edit-patient.html", data)
		return
	}
	if r.PostFormValue("Ischild") == "true" {
		child = true
	} else {
		child = false
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
		About:           "",
		Verified:        false,
		Ischild:         child,
		Hashed_password: hashed_password,
		Created_at:      time.Now(),
	}
	if _, err := server.Services.PatientService.Create(patient); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Exists"] = err.Error()
		data.Errors = msg.Errors
		server.Templates.Render(w, "admin-edit-patient.html", data)
		return
	}
	http.Redirect(w, r, "/admin/home", 300)
}

func (server *Server) Admincreateschedule(w http.ResponseWriter, r *http.Request) {
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" || v == "editor" || v == "admin:admin" || v == "admin:editor" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
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
		User   UserResp
		Active []string
		Errors Errors
		Csrf   map[string]interface{}
	}{
		Active: active_inactive(),
		User:   admin,
		Errors: msg.Errors,
		Csrf:   msg.Csrf,
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "admin-edit-schedule.html", data)
		return
	}
	if ok := msg.Validate(); !ok {
		data.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-edit-schedule.html", data)
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
		server.Templates.Render(w, "admin-edit-schedule.html", data)
		return
	}
	http.Redirect(w, r, "/admin/home", 300)
}

func (server *Server) AdmincreateAppointment(w http.ResponseWriter, r *http.Request) {
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" || v == "editor" || v == "appointment:admin" || v == "appointment:editor" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}
	var msg Form
	var approval bool
	register := Appointment{
		Doctorid:        r.PostFormValue("Doctorid"),
		Patientid:       r.PostFormValue("Patientid"),
		AppointmentDate: r.PostFormValue("Appointmentdate"),
		Duration:        r.PostFormValue("Duration"),
		Approval:        r.PostFormValue("Approval"),
	}
	msg = NewForm(r, &register)

	data := struct {
		User     UserResp
		Errors   Errors
		Approval []string
		Csrf     map[string]interface{}
	}{
		User:     admin,
		Errors:   msg.Errors,
		Approval: active_inactive(),
		Csrf:     msg.Csrf,
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "admin-edit-apntmt.html", data)
		return
	}
	if ok := msg.Validate(); !ok {
		data.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-edit-apntmt.html", data)
		return
	}

	doctorid, _ := strconv.Atoi(register.Doctorid)
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
		Doctorid:        doctorid,
		Patientid:       patientid,
		Appointmentdate: date,
		Duration:        register.Duration,
		Approval:        approval,
	}
	_, err = server.Services.DoctorBookAppointment(apntmt)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Exists"] = err.Error()
		data.Errors = msg.Errors
		server.Templates.Render(w, "admin-edit-apntmt.html", data)
		return
	}
	http.Redirect(w, r, "/admin/home", 300)
}

func (server *Server) Admincreaterecords(w http.ResponseWriter, r *http.Request) {
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" || v == "editor" || v == "records:admin" || v == "records:editor" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}
	var msg Form
	doctorid, _ := strconv.Atoi(r.PostFormValue("Doctorid"))
	patientid, _ := strconv.Atoi(r.PostFormValue("Patientid"))

	register := Records{
		Patientid:    r.PostFormValue("Doctorid"),
		Doctorid:     r.PostFormValue("Doctorid"),
		Diagnosis:    r.PostFormValue("Diagnosis"),
		Disease:      r.PostFormValue("Disease"),
		Prescription: r.PostFormValue("Prescription"),
		Weight:       r.PostFormValue("Weight"),
	}
	msg = NewForm(r, &register)

	data := struct {
		User   UserResp
		Errors Errors
		Csrf   map[string]interface{}
	}{
		User:   admin,
		Errors: msg.Errors,
		Csrf:   msg.Csrf,
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "admin-edit-records.html", data)
		return
	}
	if ok := msg.Validate(); !ok {
		data.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-edit-records.html", data)
		return
	}
	records := models.Patientrecords{
		Patienid: patientid,
		Doctorid: doctorid,
		// Diagnosis:    register.Diagnosis,
		// Disease:      register.Diagnosis,
		// Prescription: register.Prescription,
		Weight: register.Weight,
		Date:   time.Now(),
	}
	if _, err := server.Services.PatientRecordService.Create(records); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Exists"] = "record already exist"
		data.Errors = msg.Errors
		server.Templates.Render(w, "admin-edit-records.html", data)
		return
	}
	http.Redirect(w, r, "/admin/home", 300)
}

func (server *Server) Admincreatedepartment(w http.ResponseWriter, r *http.Request) {
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" || v == "editor" || v == "department:admin" || v == "department:editor" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}
	var msg Form
	register := Department{
		Departmentname: r.PostFormValue("Departmentname"),
	}
	msg = NewForm(r, &register)
	data := struct {
		User   UserResp
		Errors Errors
		Csrf   map[string]interface{}
	}{
		User:   admin,
		Errors: msg.Errors,
		Csrf:   msg.Csrf,
	}

	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "admin-edit-department.html", data)
		return
	}
	if ok := msg.Validate(); !ok {
		data.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-edit-department.html", data)
		return
	}
	// dob, _ := time.Parse("2006-01-02", register.Dob)
	dept := models.Department{
		Departmentname: register.Departmentname,
	}
	if _, err := server.Services.DepartmentService.Create(dept); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Exists"] = "department already exists"
		data.Errors = msg.Errors
		server.Templates.Render(w, "admin-edit-department.html", data)
		return
	}
	http.Redirect(w, r, "/admin/home", 300)
}

func (server *Server) Admincreatedoctor(w http.ResponseWriter, r *http.Request) {
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" || v == "editor" || v == "physician:admin" || v == "physician:editor" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}
	var msg Form
	register := DocRegister{
		Email:           r.PostFormValue("Email"),
		Password:        r.PostFormValue("Password"),
		ConfirmPassword: r.PostFormValue("ConfirmPassword"),
		Username:        r.PostFormValue("Username"),
		Fullname:        r.PostFormValue("Fullname"),
		Contact:         r.PostFormValue("Contact"),
		Departmentname:  r.PostFormValue("Departmentname"),
	}
	msg = NewForm(r, &register)
	data := struct {
		User   UserResp
		Errors Errors
		Csrf   map[string]interface{}
	}{
		User:   admin,
		Errors: msg.Errors,
		Csrf:   msg.Csrf,
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "admin-edit-doctor.html", data)
		return
	}
	if ok := msg.Validate(); !ok {
		data.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-edit-doctor.html", data)
		return
	}
	// dob, _ := time.Parse("2006-01-02", register.Dob)
	hashed_password, _ := services.HashPassword(register.Password)
	doctor := models.Physician{
		Username:        register.Username,
		Full_name:       register.Fullname,
		Email:           register.Email,
		Contact:         register.Contact,
		Hashed_password: hashed_password,
		About:           " ",
		Verified:        false,
		Departmentname:  register.Departmentname,
		Created_at:      time.Now(),
	}
	if _, err := server.Services.DoctorService.Create(doctor); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Exists"] = "doctor already exists"
		data.Errors = msg.Errors
		server.Templates.Render(w, "admin-edit-doctor.html", data)
		return
	}
	http.Redirect(w, r, "/admin/home", 300)
}
func (server *Server) Admincreatenurse(w http.ResponseWriter, r *http.Request) {
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" || v == "editor" || v == "nurse:admin" || v == "nurse:editor" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}
	var msg Form
	register := NurseRegister{
		Email:           r.PostFormValue("Email"),
		Password:        r.PostFormValue("Password"),
		ConfirmPassword: r.PostFormValue("ConfirmPassword"),
		Username:        r.PostFormValue("Username"),
		Fullname:        r.PostFormValue("Fullname"),
	}
	msg = NewForm(r, &register)
	data := struct {
		User   UserResp
		Errors Errors
		Csrf   map[string]interface{}
	}{
		User:   admin,
		Errors: msg.Errors,
		Csrf:   msg.Csrf,
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "admin-edit-nurse.html", data)
		return
	}
	if ok := msg.Validate(); !ok {
		data.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-edit-nurse.html", data)
		return
	}
	// dob, _ := time.Parse("2006-01-02", register.Dob)
	hashed_password, _ := services.HashPassword(register.Password)
	nurse := models.Nurse{
		Username:        register.Username,
		Full_name:       register.Fullname,
		Email:           register.Email,
		Hashed_password: hashed_password,
		Created_at:      time.Now(),
	}
	if _, err := server.Services.NurseService.Create(nurse); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Exists"] = "doctor already exists"
		data.Errors = msg.Errors
		server.Templates.Render(w, "admin-edit-nurse.html", data)
		return
	}
	http.Redirect(w, r, "/admin/home", 300)
}

func (server *Server) Admindeletedoctor(w http.ResponseWriter, r *http.Request) {
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" || v == "physician:admin" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := server.Services.DoctorService.Delete(idparam); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/500", 300)
	}
	http.Redirect(w, r, "/admin/home", 300)
}

func (server *Server) Admindeletepatient(w http.ResponseWriter, r *http.Request) {
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" || v == "patient:admin" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}

	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := server.Services.PatientService.Delete(idparam); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-edit-patient.html", nil)
		return
	}
	http.Redirect(w, r, "/admin/home", 300)
}
func (server *Server) Admindeletedepartment(w http.ResponseWriter, r *http.Request) {
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" || v == "department:admin" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := server.Services.DepartmentService.Delete(idparam); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-edit-department.html", nil)
		return
	}
	http.Redirect(w, r, "/admin/home", 300)
}

func (server *Server) Admindeleterecord(w http.ResponseWriter, r *http.Request) {
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" || v == "records:admin" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}

	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := server.Services.PatientRecordService.Delete(idparam); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-edit-records.html", nil)
		return
	}
	http.Redirect(w, r, "/admin/home", 300)
}

func (server *Server) Admindeleteappointment(w http.ResponseWriter, r *http.Request) {
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" || v == "appointment:admin" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}

	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := server.Services.AppointmentService.Delete(idparam); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-edit-apntmt.html", nil)
		return
	}
	http.Redirect(w, r, "/admin/home", 300)
}

func (server *Server) Admindeleteschedule(w http.ResponseWriter, r *http.Request) {
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" || v == "schedule:admin" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
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
		server.Templates.Render(w, "admin-edit-schedule.html", nil)
		return
	}
	http.Redirect(w, r, "/admin/home", 300)
}

func active_inactive() []string {
	var status = []string{
		"Active",
		"Inactive",
	}
	return status
}
func (server *Server) Adminupdatepatient(w http.ResponseWriter, r *http.Request) {
	var msg Form
	var child bool
	Errmap := make(map[string]string)
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := server.Services.PatientService.Find(idparam)
	if err != nil {
		server.Templates.Render(w, "404.html", nil)
		return
	}
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" || v == "editor" || v == "patient:admin" || v == "patient:editor" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
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
	msg = NewForm(r, &register)
	pdata := struct {
		User       UserResp
		Errors     Errors
		Patient    models.Patient
		Bloodgroup []string
		Csrf       map[string]interface{}
	}{
		Errors:     Errmap,
		Patient:    data,
		User:       admin,
		Bloodgroup: bloodgroup_array(),
		Csrf:       msg.Csrf,
	}
	if r.PostFormValue("Ischild") == "true" {
		child = true
	} else {
		child = false
	}
	dob, _ := time.Parse("2006-01-02", register.Dob)
	hashed_password, _ := services.HashPassword(register.Password)
	patient := models.Patient{
		Patientid:       idparam,
		Username:        register.Username,
		Full_name:       register.Fullname,
		Email:           register.Email,
		Dob:             dob,
		Contact:         register.Contact,
		Verified:        false,
		About:           "",
		Ischild:         child,
		Bloodgroup:      register.Bloodgroup,
		Hashed_password: hashed_password,
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "admin-update-patient.html", pdata)
		return
	}
	if ok := msg.Validate(); !ok {
		pdata.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-update-patient.html", pdata)
		return
	}

	dt := struct {
		User       UserResp
		Errors     Errors
		Bloodgroup []string
		Csrf       map[string]interface{}
	}{
		User:       admin,
		Errors:     Errmap,
		Bloodgroup: bloodgroup_array(),
		Csrf:       msg.Csrf,
	}
	if _, err := server.Services.PatientService.Update(patient); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		dt.Errors = Errmap
		server.Templates.Render(w, "admin-update-patient.html", dt)
		return
	}
	http.Redirect(w, r, r.URL.String(), 301)
}

func (server *Server) Adminupdateschedule(w http.ResponseWriter, r *http.Request) {
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" || v == "editor" || v == "schedule:admin" || v == "schedule:editor" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
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
		User     UserResp
		Errors   Errors
		Csrf     map[string]interface{}
		Schedule models.Schedule
		Active   []string
	}{
		Errors:   Errmap,
		Schedule: data,
		Csrf:     msg.Csrf,
		User:     admin,
		Active:   active_inactive(),
	}
	var actvie bool
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "admin-update-schedule.html", pdata)
		return
	}
	if ok := msg.Validate(); !ok {
		pdata.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-update-schedule.html", pdata)
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
		User   UserResp
		Csrf   map[string]interface{}
		Errors Errors
		Active []string
	}{
		User:   admin,
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
		server.Templates.Render(w, "admin-update-schedule.html", dt)
		return
	}
	http.Redirect(w, r, r.URL.String(), 301)
}

func (server *Server) AdminupdateAppointment(w http.ResponseWriter, r *http.Request) {
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" || v == "editor" || v == "appointment:admin" || v == "appointment:editor" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
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
		User        UserResp
		Errors      Errors
		Csrf        map[string]interface{}
		Appointment models.Appointment
		Approval    []string
	}{
		Errors:      Errmap,
		Appointment: data,
		User:        admin,
		Approval:    active_inactive(),
		Csrf:        msg.Csrf,
	}
	var approval bool
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "admin-update-appointment.html", pdata)
		return
	}
	if ok := msg.Validate(); !ok {
		pdata.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-update-appointment.html", pdata)
		return
	}

	dt := struct {
		User   UserResp
		Csrf   map[string]interface{}
		Errors Errors
	}{
		User:   admin,
		Csrf:   msg.Csrf,
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

	if _, err := server.Services.UpdateappointmentbyDoctor(apntmt); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		dt.Errors = Errmap
		server.Templates.Render(w, "admin-update-appointment.html", dt)
		return
	}
	http.Redirect(w, r, r.URL.String(), 301)
}

func (server *Server) Adminupdaterecords(w http.ResponseWriter, r *http.Request) {
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" || v == "editor" || v == "records:admin" || v == "records:editor" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
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
	pdata := struct {
		User    UserResp
		Errors  Errors
		Records models.Patientrecords
		Csrf    map[string]interface{}
	}{
		Errors:  Errmap,
		Records: data,
		User:    admin,
		Csrf:    msg.Csrf,
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "admin-update-record.html", pdata)
		return
	}
	if ok := msg.Validate(); !ok {
		pdata.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-update-record.html", pdata)
		return
	}

	dt := struct {
		User   UserResp
		Errors Errors
		Csrf   map[string]interface{}
	}{
		User:   admin,
		Errors: Errmap,
		Csrf:   msg.Csrf,
	}
	doctorid, _ := strconv.Atoi(r.PostFormValue("Doctorid"))
	patientid, _ := strconv.Atoi(r.PostFormValue("Patientid"))
	records := models.Patientrecords{
		Recordid: data.Recordid,
		Patienid: patientid,
		Doctorid: doctorid,
		// Diagnosis:    register.Diagnosis,
		// Disease:      register.Diagnosis,
		// Prescription: register.Prescription,
		Weight: register.Weight,
		Date:   time.Now(),
	}

	if _, err := server.Services.PatientRecordService.Update(records); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		dt.Errors = Errmap
		server.Templates.Render(w, "admin-update-record.html", dt)
		return
	}
	http.Redirect(w, r, r.URL.String(), 301)
}

func (server *Server) Adminupdatenurse(w http.ResponseWriter, r *http.Request) {
	var msg Form
	Errmap := make(map[string]string)
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := server.Services.NurseService.Find(idparam)
	if err != nil {
		server.Templates.Render(w, "404.html", nil)
	}
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" || v == "editor" || v == "nurse:admin" || v == "nurse:editor" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}

	register := NurseRegister{
		Email:           r.PostFormValue("Email"),
		Password:        r.PostFormValue("Password"),
		ConfirmPassword: r.PostFormValue("ConfirmPassword"),
		Username:        r.PostFormValue("Username"),
		Fullname:        r.PostFormValue("Fullname"),
	}

	msg = NewForm(r, &register)
	pdata := struct {
		User   UserResp
		Errors Errors
		Nurse  models.Nurse
		Csrf   map[string]interface{}
	}{
		Errors: Errmap,
		Nurse:  data,
		User:   admin,
		Csrf:   msg.Csrf,
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "admin-update-nurse.html", pdata)
		return
	}
	if ok := msg.Validate(); !ok {
		pdata.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-update-nurse.html", pdata)
		return
	}

	dt := struct {
		User   UserResp
		Errors Errors
		Csrf   map[string]interface{}
	}{
		User:   admin,
		Errors: Errmap,
		Csrf:   msg.Csrf,
	}
	hashed_password, _ := services.HashPassword(register.Password)
	nurse := models.Nurse{
		Id:              data.Id,
		Username:        register.Username,
		Full_name:       register.Fullname,
		Email:           register.Email,
		Hashed_password: hashed_password,
	}
	if _, err := server.Services.NurseService.Update(nurse); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		dt.Errors = Errmap
		server.Templates.Render(w, "admin-update-nurse.html", dt)
		return
	}
	http.Redirect(w, r, r.URL.String(), 301)
}
func (server *Server) Adminupdatedoctor(w http.ResponseWriter, r *http.Request) {
	var msg Form
	Errmap := make(map[string]string)
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := server.Services.DoctorService.Find(idparam)
	if err != nil {
		server.Templates.Render(w, "404.html", nil)
	}
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" || v == "editor" || v == "physician:admin" || v == "physician:editor" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}

	// var approval bool
	register := DocRegister{
		Email:           r.PostFormValue("Email"),
		Password:        r.PostFormValue("Password"),
		ConfirmPassword: r.PostFormValue("ConfirmPassword"),
		Username:        r.PostFormValue("Username"),
		Fullname:        r.PostFormValue("Fullname"),
		Contact:         r.PostFormValue("Contact"),
		Departmentname:  r.PostFormValue("Departmentname"),
	}
	msg = NewForm(r, &register)
	pdata := struct {
		User   UserResp
		Errors Errors
		Doctor models.Physician
		Csrf   map[string]interface{}
	}{
		Errors: Errmap,
		Doctor: data,
		User:   admin,
		Csrf:   msg.Csrf,
	}
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "admin-update-doctor.html", pdata)
		return
	}
	if ok := msg.Validate(); !ok {
		pdata.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-update-doctor.html", pdata)
		return
	}

	dt := struct {
		User   UserResp
		Errors Errors
		Csrf   map[string]interface{}
	}{
		User:   admin,
		Errors: Errmap,
		Csrf:   msg.Csrf,
	}
	hashed_password, _ := services.HashPassword(register.Password)
	doctor := models.Physician{
		Physicianid:     data.Physicianid,
		Username:        register.Username,
		Full_name:       register.Fullname,
		Email:           register.Email,
		Contact:         register.Contact,
		Hashed_password: hashed_password,
		Departmentname:  register.Departmentname,
	}
	if _, err := server.Services.DoctorService.Update(doctor); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		dt.Errors = Errmap
		server.Templates.Render(w, "admin-update-doctor.html", dt)
		return
	}
	http.Redirect(w, r, r.URL.String(), 301)
}

func (server *Server) Adminupdatedepartment(w http.ResponseWriter, r *http.Request) {
	var msg Form
	Errmap := make(map[string]string)
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := server.Services.DepartmentService.Find(idparam)
	if err != nil {
		server.Templates.Render(w, "404.html", nil)
	}
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
	var acceptedperm []string
	for _, v := range admin.Permission {
		if v == "admin" || v == "editor" || v == "department:admin" || v == "department:editor" {
			acceptedperm = append(acceptedperm, v)
		}
	}
	if acceptedperm == nil {
		w.WriteHeader(http.StatusUnauthorized)
		server.Templates.Render(w, "401.html", nil)
		return
	}

	pdata := struct {
		User       UserResp
		Errors     Errors
		Department models.Department
		Csrf       map[string]interface{}
	}{
		Errors:     Errmap,
		Department: data,
		User:       admin,
		Csrf:       msg.Csrf,
	}
	register := Department{
		Departmentname: r.PostFormValue("Departmentname"),
	}
	msg = NewForm(r, &register)
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		server.Templates.Render(w, "admin-update-dept.html", pdata)
		return
	}
	if ok := msg.Validate(); !ok {
		pdata.Errors = msg.Errors
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-update-dept.html", pdata)
		return
	}

	dt := struct {
		User   UserResp
		Errors Errors
		Csrf   map[string]interface{}
	}{
		User:   admin,
		Errors: Errmap,
		Csrf:   msg.Csrf,
	}
	dept := models.Department{
		Departmentid:   data.Departmentid,
		Departmentname: register.Departmentname,
	}
	if _, err := server.Services.DepartmentService.Update(dept); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		dt.Errors = Errmap
		server.Templates.Render(w, "admin-update-dept.html", dt)
		return
	}
	http.Redirect(w, r, r.URL.String(), 301)
}
