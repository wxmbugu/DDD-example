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
	session, _ := server.Store.Get(r, "admin")
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
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
		return
	}
	if err = services.CheckPassword(user.Password, login.Password); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Login"] = "No such user"
		server.Templates.Render(w, "login.html", msg)
		return
	}
	permission, err := server.Services.RbacService.PermissionsService.FindbyRoleId(user.Roleid)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
		return
	}
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
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
		return
	}
	http.Redirect(w, r, "/admin/home", http.StatusMovedPermanently)
}

func (server *Server) AdminLogout(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	session.Values["admin"] = UserResp{}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	http.Redirect(w, r, "/admin/home", http.StatusMovedPermanently)
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
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
		return
	}
	user := getAdmin(session)
	if !user.Authenticated {
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}
	_, ametadata, err := server.Services.AppointmentService.FindAll(models.Filters{
		PageSize: 20,
		Page:     1,
	})
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	_, rmetadata, err := server.Services.PatientRecordService.FindAll(
		models.Filters{
			PageSize: 20,
			Page:     1,
		})
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	data := struct {
		User    UserResp
		Apntmt  int
		Records int
	}{
		User:    user,
		Apntmt:  ametadata.TotalRecords,
		Records: rmetadata.TotalRecords,
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "admin-home.html", data)
}

type Pagination struct {
	Page      int
	LastPage  int
	PrevPage  int
	NextPage  int
	HasPrev   bool
	HasNext   bool
	Count     int
	FirstPage int
}

func Newpagination(metadata models.Metadata) Pagination {
	return Pagination{
		Count:     metadata.TotalRecords,
		LastPage:  metadata.LastPage,
		FirstPage: metadata.FirstPage,
	}
}
func (p *Pagination) nextpage(id int) {
	if p.Count <= id*PageCount {
		p.Page = id
		p.HasNext = false
	} else {
		p.HasNext = true
		p.NextPage = id + 1
		p.Page = id
	}
}

func (p *Pagination) previouspage(id int) {
	if id <= 1 {
		p.HasPrev = false
		p.PrevPage = 1
		p.Page = id
	} else {
		p.HasPrev = true
		p.PrevPage = id - 1
		p.Page = id
	}
}
func (server *Server) Adminrecord(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
		return
	}
	params := mux.Vars(r)
	id := params["pageid"]
	idparam, err := strconv.Atoi(id)
	if err != nil || idparam <= 0 {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	records, metadata, err := server.Services.PatientRecordService.FindAll(
		models.Filters{
			PageSize: PageCount,
			Page:     idparam,
		})
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	paging := Newpagination(*metadata)
	paging.nextpage(idparam)
	paging.previouspage(idparam)
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
}

func (server *Server) Adminappointments(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	}
	params := mux.Vars(r)
	id := params["pageid"]
	idparam, err := strconv.Atoi(id)
	if err != nil || idparam <= 0 {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	appointment, metadata, err := server.Services.AppointmentService.FindAll(models.Filters{
		PageSize: PageCount,
		Page:     idparam,
	})
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	paging := Newpagination(*metadata)
	paging.nextpage(idparam)
	paging.previouspage(idparam)
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
}

func (server *Server) Adminuser(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	}
	users, err := server.Services.RbacService.UsersService.FindAll()
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	data := struct {
		User  UserResp
		Users []models.Users
	}{
		User:  admin,
		Users: users,
	}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "admin-user.html", data)
}

func (server *Server) Admincreateuser(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
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
		User    UserResp
		Errors  Errors
		Csrf    map[string]interface{}
		Success string
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
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	password, err := services.HashPassword(register.Password)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
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
	w.WriteHeader(http.StatusCreated)
	data.Success = "Created account successfully"
	server.Templates.Render(w, "admin-edit-user.html", data)
}

func (server *Server) Adminupdateuser(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
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
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	role, err := server.Services.RbacService.RolesService.Find(user.Roleid)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	data := struct {
		User      UserResp
		Errors    Errors
		Csrf      map[string]interface{}
		AdminUser models.Users
		Role      string
		Success   string
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
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	password, err := services.HashPassword(register.Password)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	user, err = server.Services.RbacService.UsersService.Update(models.Users{
		Id:       user.Id,
		Email:    register.Email,
		Password: password,
		Roleid:   role.Roleid,
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg.Errors["Exists"] = err.Error()
		data.Errors = msg.Errors
		server.Templates.Render(w, "admin-update-user.html", data)
		return
	}
	w.WriteHeader(http.StatusOK)
	data.AdminUser = user
	data.Role = role.Role
	data.Success = "account updated successfully"
	server.Templates.Render(w, "admin-update-user.html", data)
}

func (server *Server) Admindeleteuser(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	}
	if err := server.Services.RbacService.UsersService.Delete(idparam); err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
}

func (server *Server) Admindeleterole(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}
	if err := server.Services.RbacService.RolesService.Delete(idparam); err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
}
func (server *Server) Admindeletenurse(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	}
	if err := server.Services.NurseService.Delete(idparam); err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
}
func (server *Server) Adminroles(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	}
	roles, err := server.Services.RbacService.RolesService.FindAll()
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
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
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
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
		Success    string
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
	w.WriteHeader(http.StatusCreated)
	data.Success = "role created successfully"
	server.Templates.Render(w, "admin-edit-role.html", data)
}

func (server *Server) Adminupdateroles(w http.ResponseWriter, r *http.Request) {
	available_permissions := generate_permission()
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
	}
	var msg Form
	assigned_permissions, err := server.Services.RbacService.PermissionsService.FindbyRoleId(idparam)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
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
	http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
}

func (server *Server) Adminfilterphysician(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	name := r.URL.Query().Get("name")
	dept := r.URL.Query().Get("dept")
	admin := getAdmin(session)
	if !admin.Authenticated {
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
		if filtermap["name"] != "" && filtermap["dept"] != "" {
			name = filtermap["name"]
			dept = filtermap["dept"]
			url := r.URL.Path + `?pageid=1` + "&" + "name=" + name + "&" + "dept=" + dept
			http.Redirect(w, r, url, http.StatusMovedPermanently)

		} else if filtermap["name"] == "" && filtermap["dept"] != "" {
			dept = filtermap["dept"]
			url := r.URL.Path + `?pageid=1` + "&" + "dept=" + dept
			http.Redirect(w, r, url, http.StatusMovedPermanently)
		} else if filtermap["name"] != "" && filtermap["dept"] == "" {
			name = filtermap["name"]
			url := r.URL.Path + `?pageid=1` + "&" + "name=" + name
			http.Redirect(w, r, url, http.StatusMovedPermanently)
		}
	}
	id := r.URL.Query().Get("pageid")
	idparam, err := strconv.Atoi(id)
	if err != nil || idparam <= 0 {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	doctors, metadata, err := server.Services.DoctorService.Filter(name, dept, models.Filters{
		PageSize: PageCount,
		Page:     idparam,
	})
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	paging := Newpagination(*metadata)
	paging.nextpage(idparam)
	paging.previouspage(idparam)
	data := struct {
		User       UserResp
		Doctors    []*models.Physician
		Pagination Pagination
		Csrf       map[string]interface{}
	}{
		User:       admin,
		Doctors:    doctors,
		Pagination: paging,
		Csrf:       form.Csrf}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "admin-physician.html", data)
}

func (server *Server) Adminfilterpatient(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	name := r.URL.Query().Get("name")
	admin := getAdmin(session)
	if !admin.Authenticated {
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
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	patient, metadata, err := server.Services.PatientService.Filter(name, models.Filters{
		PageSize: PageCount,
		Page:     idparam,
	})
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	paging := Newpagination(*metadata)
	paging.nextpage(idparam)
	paging.previouspage(idparam)
	data := struct {
		User       UserResp
		Patient    []*models.Patient
		Pagination Pagination
		Csrf       map[string]interface{}
	}{
		User:       admin,
		Patient:    patient,
		Pagination: paging,
		Csrf:       form.Csrf}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "admin-patient.html", data)
}
func (server *Server) Adminfilternurse(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	name := r.URL.Query().Get("name")
	admin := getAdmin(session)
	if !admin.Authenticated {
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
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	nurse, metadata, err := server.Services.NurseService.Filter(name, models.Filters{
		PageSize: PageCount,
		Page:     idparam,
	})
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	paging := Newpagination(*metadata)
	paging.nextpage(idparam)
	paging.previouspage(idparam)
	data := struct {
		User       UserResp
		Nurse      []*models.Nurse
		Pagination Pagination
		Csrf       map[string]interface{}
	}{
		User:       admin,
		Nurse:      nurse,
		Pagination: paging,
		Csrf:       form.Csrf}
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "admin-nurse.html", data)
}

func (server *Server) Adminschedule(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
	}
	params := mux.Vars(r)
	id := params["pageid"]
	idparam, err := strconv.Atoi(id)
	if err != nil || idparam <= 0 {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	schedules, metadata, err := server.Services.ScheduleService.FindAll(models.Filters{
		PageSize: PageCount,
		Page:     idparam,
	})
	paging := Newpagination(*metadata)
	paging.nextpage(idparam)
	paging.previouspage(idparam)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
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
}

func (server *Server) Admindepartment(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
	}
	params := mux.Vars(r)
	id := params["pageid"]
	idparam, err := strconv.Atoi(id)
	if err != nil || idparam <= 0 {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	department, metadata, err := server.Services.DepartmentService.FindAll(models.Filters{
		PageSize: PageCount,
		Page:     idparam,
	})
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	paging := Newpagination(*metadata)
	paging.nextpage(idparam)
	paging.previouspage(idparam)
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
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
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
		Success    string
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
	w.WriteHeader(http.StatusCreated)
	data.Success = "account created successfully"
	server.Templates.Render(w, "admin-edit-patient.html", data)
}

func (server *Server) Admincreateschedule(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
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
		User    UserResp
		Active  []string
		Errors  Errors
		Csrf    map[string]interface{}
		Success string
	}{
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
		server.Templates.Render(w, "admin-edit-schedule.html", data)
		return
	}
	w.WriteHeader(http.StatusCreated)
	data.Success = "schedule created succcessfully"
	server.Templates.Render(w, "admin-edit-schedule.html", data)
}

func (server *Server) AdmincreateAppointment(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
	}
	var msg Form
	var approval bool
	register := Appointment{
		Doctorid:        r.PostFormValue("Doctorid"),
		Patientid:       r.PostFormValue("Patientid"),
		AppointmentDate: r.PostFormValue("Appointmentdate"),
		Duration:        r.PostFormValue("Duration"),
	}
	msg = NewForm(r, &register)

	data := struct {
		User     UserResp
		Errors   Errors
		Approval []string
		Csrf     map[string]interface{}
		Success  string
	}{
		User:   admin,
		Errors: msg.Errors,
		Csrf:   msg.Csrf,
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
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	approval = checkboxvalue(r.PostFormValue("Approval"))
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
	w.WriteHeader(http.StatusCreated)
	data.Success = "appointment created successfully"
	server.Templates.Render(w, "admin-edit-apntmt.html", data)
}

func (server *Server) Admincreaterecords(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
	}
	var msg Form
	height, _ := strconv.Atoi(r.PostFormValue("Height"))
	bp, _ := strconv.Atoi(r.PostFormValue("Bp"))
	temp, _ := strconv.Atoi(r.PostFormValue("Temperature"))
	patientid, _ := strconv.Atoi(r.PostFormValue("Patientid"))
	doctorid, _ := strconv.Atoi(r.PostFormValue("Doctorid"))
	nurseid, _ := strconv.Atoi(r.PostFormValue("Nurseid"))
	hr, _ := strconv.Atoi(r.PostFormValue("HeartRate"))
	register := Records{
		Height:      r.PostFormValue("Height"),
		Bp:          r.PostFormValue("Bp"),
		Temperature: r.PostFormValue("Temperature"),
		Weight:      r.PostFormValue("Weight"),
		Patientid:   r.PostFormValue("Patientid"),
		HeartRate:   r.PostFormValue("HeartRate"),
		Doctorid:    r.PostFormValue("Doctorid"),
	}
	msg = NewForm(r, &register)

	data := struct {
		User    UserResp
		Errors  Errors
		Csrf    map[string]interface{}
		Success string
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
		Doctorid:    doctorid,
		Patienid:    patientid,
		Nurseid:     nurseid,
		Height:      height,
		HeartRate:   hr,
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
		server.Templates.Render(w, "admin-edit-records.html", data)
		return
	}
	w.WriteHeader(http.StatusCreated)
	data.Success = "record created"
	server.Templates.Render(w, "admin-edit-records.html", data)
}

func (server *Server) Admincreatedepartment(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
	}
	var msg Form
	register := Department{
		Departmentname: r.PostFormValue("Departmentname"),
	}
	msg = NewForm(r, &register)
	data := struct {
		User    UserResp
		Errors  Errors
		Csrf    map[string]interface{}
		Success string
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
	w.WriteHeader(http.StatusCreated)
	data.Success = "department created successfully"
	server.Templates.Render(w, "admin-edit-department.html", data)
}

func (server *Server) Admincreatedoctor(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
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
		User    UserResp
		Errors  Errors
		Csrf    map[string]interface{}
		Success string
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
	w.WriteHeader(http.StatusCreated)
	data.Success = "accounted created successfully"
	server.Templates.Render(w, "admin-edit-doctor.html", data)
}
func (server *Server) Admincreatenurse(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
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
		User    UserResp
		Errors  Errors
		Csrf    map[string]interface{}
		Success string
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
	w.WriteHeader(http.StatusOK)
	data.Success = "account created successfully"
	server.Templates.Render(w, "admin-edit-nurse.html", data)
}

func (server *Server) Admindeletedoctor(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
	}
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	if err := server.Services.DoctorService.Delete(idparam); err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
}

func (server *Server) Admindeletepatient(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
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
	http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
}
func (server *Server) Admindeletedepartment(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
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
	http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
}

func (server *Server) Admindeleterecord(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
	}
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	if err := server.Services.PatientRecordService.Delete(idparam); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-edit-records.html", nil)
		return
	}
	http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
}

func (server *Server) Admindeleteappointment(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
	}
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	if err := server.Services.AppointmentService.Delete(idparam); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-edit-apntmt.html", nil)
		return
	}
	http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
}

func (server *Server) Admindeleteschedule(w http.ResponseWriter, r *http.Request) {
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
	}
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}

	if err := server.Services.ScheduleService.Delete(idparam); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		server.Templates.Render(w, "admin-edit-schedule.html", nil)
		return
	}
	http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
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
		w.WriteHeader(http.StatusNotFound)
		server.Templates.Render(w, "404.html", nil)
		return
	}
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
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
		Success    string
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
	if _, err := server.Services.PatientService.Update(patient); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		pdata.Errors = Errmap
		server.Templates.Render(w, "admin-update-patient.html", pdata)
		return
	}
	w.WriteHeader(http.StatusOK)
	pdata.Patient = patient
	pdata.Success = "accounted updated successfully"
	server.Templates.Render(w, "admin-update-patient.html", pdata)
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
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
	}
	register := Schedule{
		Doctorid:  r.PostFormValue("Doctorid"),
		Starttime: r.PostFormValue("Starttime"),
		Endtime:   r.PostFormValue("Endtime"),
	}
	msg = NewForm(r, &register)
	pdata := struct {
		User     UserResp
		Errors   Errors
		Csrf     map[string]interface{}
		Schedule models.Schedule
		Success  string
	}{
		Errors:   Errmap,
		Schedule: data,
		Csrf:     msg.Csrf,
		User:     admin,
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
	actvie = checkboxvalue(r.PostFormValue("Active"))
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
		pdata.Errors = Errmap
		server.Templates.Render(w, "admin-update-schedule.html", pdata)
		return
	}
	w.WriteHeader(http.StatusOK)
	pdata.Schedule = schedule
	pdata.Success = "schedule updated successfully"
	server.Templates.Render(w, "admin-update-schedule.html", pdata)
}

func (server *Server) AdminupdateAppointment(w http.ResponseWriter, r *http.Request) {
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
		server.Templates.Render(w, "404.html", nil)
	}
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
	}
	register := Appointment{
		Doctorid:        r.PostFormValue("Doctorid"),
		Patientid:       r.PostFormValue("Patientid"),
		AppointmentDate: r.PostFormValue("Appointmentdate"),
		Duration:        r.PostFormValue("Duration"),
	}
	msg = NewForm(r, &register)
	pdata := struct {
		User        UserResp
		Errors      Errors
		Csrf        map[string]interface{}
		Appointment models.Appointment
		Success     string
	}{
		Errors:      Errmap,
		Appointment: data,
		User:        admin,
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
	doctorid, _ := strconv.Atoi(r.PostFormValue("Doctorid"))
	patientid, _ := strconv.Atoi(r.PostFormValue("Patientid"))
	date, err := time.Parse("2006-01-02T15:04", r.PostFormValue("Appointmentdate"))
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	approval = checkboxvalue(r.PostFormValue("Approval"))
	outbound := checkboxvalue(r.PostFormValue("Outbound"))
	apntmt := models.Appointment{
		Appointmentid:   data.Appointmentid,
		Doctorid:        doctorid,
		Patientid:       patientid,
		Appointmentdate: date,
		Duration:        register.Duration,
		Approval:        approval,
		Outbound:        outbound,
	}
	if _, err := server.Services.UpdateappointmentbyDoctor(apntmt); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		pdata.Errors = Errmap
		server.Templates.Render(w, "admin-update-appointment.html", pdata)
		return
	}
	w.WriteHeader(http.StatusOK)
	pdata.Appointment = apntmt
	pdata.Success = "appointment updated successfully"
	server.Templates.Render(w, "admin-update-appointment.html", pdata)
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
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
	}
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
	w.WriteHeader(http.StatusOK)
	server.Templates.Render(w, "admin-update-record.html", pdata)
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
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
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
		User    UserResp
		Errors  Errors
		Nurse   models.Nurse
		Csrf    map[string]interface{}
		Success string
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
		pdata.Errors = Errmap
		server.Templates.Render(w, "admin-update-nurse.html", pdata)
		return
	}
	w.WriteHeader(http.StatusOK)
	pdata.Nurse = nurse
	pdata.Success = "account updated successfully"
	server.Templates.Render(w, "admin-update-nurse.html", pdata)
}
func (server *Server) Adminupdatedoctor(w http.ResponseWriter, r *http.Request) {
	Errmap := make(map[string]string)
	params := mux.Vars(r)
	id := params["id"]
	idparam, err := strconv.Atoi(id)
	if err != nil {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
	}
	data, err := server.Services.DoctorService.Find(idparam)
	if err != nil {
		server.Templates.Render(w, "404.html", nil)
	}
	session, err := server.Store.Get(r, "admin")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
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
	pdata := struct {
		User    UserResp
		Errors  Errors
		Doctor  models.Physician
		Csrf    map[string]interface{}
		Success string
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
		pdata.Errors = Errmap
		server.Templates.Render(w, "admin-update-doctor.html", pdata)
		return
	}
	w.WriteHeader(http.StatusOK)
	pdata.Doctor = doctor
	pdata.Success = "account updated successfully"
	server.Templates.Render(w, "admin-update-doctor.html", pdata)
}

func (server *Server) Adminupdatedepartment(w http.ResponseWriter, r *http.Request) {
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
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	admin := getAdmin(session)
	if !admin.Authenticated {
		http.Redirect(w, r, "/admin/login", http.StatusMovedPermanently)
	}
	register := Department{
		Departmentname: r.PostFormValue("Departmentname"),
	}
	msg := NewForm(r, &register)
	pdata := struct {
		User       UserResp
		Errors     Errors
		Department models.Department
		Csrf       map[string]interface{}
		Success    string
	}{
		Errors:     Errmap,
		Department: data,
		User:       admin,
		Csrf:       msg.Csrf,
	}

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
	dept := models.Department{
		Departmentid:   data.Departmentid,
		Departmentname: register.Departmentname,
	}
	if _, err := server.Services.DepartmentService.Update(dept); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		pdata.Errors = Errmap
		server.Templates.Render(w, "admin-update-dept.html", pdata)
		return
	}
	w.WriteHeader(http.StatusOK)
	pdata.Department = dept
	pdata.Success = "department update successfully"
	server.Templates.Render(w, "admin-update-dept.html", pdata)
}

func (server *Server) admin_reset_password(w http.ResponseWriter, r *http.Request) {
	Errmap := make(map[string]string)
	id := r.URL.Query().Get("id")
	if !strings.Contains(id, "admin") {
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
	user, err := server.Services.RbacService.UsersService.FindbyEmail(value)
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
		User    models.Users
		Errors  Errors
		Csrf    map[string]interface{}
		Success string
	}{
		Errors: Errmap,
		User:   user,
		Csrf:   msg.Csrf,
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
	user.Password, err = services.HashPassword(register.Password)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusMovedPermanently)
	}
	if _, err := server.Services.RbacService.UsersService.Update(user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Errmap["Exists"] = err.Error()
		data.Errors = Errmap
		server.Templates.Render(w, "password_reset.html", data)
		return
	}
	w.WriteHeader(http.StatusOK)
	data.Success = "password reset succcessfully"
	server.Templates.Render(w, "password_reset.html", data)
}
