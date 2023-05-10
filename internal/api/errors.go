package api

import (
	"fmt"
	"net/http"
	"net/mail"
	"reflect"
	"time"

	"github.com/gorilla/csrf"
)

type validation interface {
	validate() (Errors, bool)
}

// Holds the error value
type Errors map[string]string

type Form struct {
	// check input type content is valid.
	Data *validation
	// Holds the error value
	Errors
	// csrf token
	Csrf map[string]interface{}
}

func NewForm(r *http.Request, v validation) Form {
	csrfmap := make(map[string]interface{})
	csrfmap[csrf.TemplateTag] = csrf.TemplateField(r)
	return Form{
		Data: &v,
		Csrf: csrfmap,
	}
}

func (f *Form) Validate() bool {
	f.Errors = make(map[string]string)
	if data, ok := validateType(*f.Data); !ok {
		f.Errors = data
		return false
	}
	return true
}
func validateType(d validation) (Errors, bool) {
	return d.validate()
}

type Login struct {
	Email    string
	Password string
	Errors
}

func (l *Login) validate() (Errors, bool) {
	l.Errors = make(map[string]string)
	l.Errors = IsEmpty(*l, l.Errors)
	if err := validateEmail(l.Email); err != nil {
		l.Errors["Email"] = "Please enter a valid email address"
	}
	if l.Password == "" {
		l.Errors["Password"] = "Password required"
	}
	return l.Errors, len(l.Errors) == 0
}

type Register struct {
	Email           string
	Password        string
	Bloodgroup      string
	Username        string
	Fullname        string
	Contact         string
	Dob             string
	ConfirmPassword string
	Errors
}

func (r *Register) validate() (Errors, bool) {
	r.Errors = make(map[string]string)
	err := validateEmail(r.Email)
	if err != nil {
		r.Errors["Email"] = "Please enter a valid email address"
	}
	if r.ConfirmPassword != r.Password {
		r.Errors["Match"] = "Password & ConfirmPassword don't match"
	}
	today := time.Now()
	dob, _ := time.Parse("2006-01-02", r.Dob)
	if dob.After(today) {
		r.Errors["DobError"] = "You can't be from the future!Check Dob"
	}
	r.Errors = IsEmpty(*r, r.Errors)

	return r.Errors, len(r.Errors) == 0
}

func validateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	return err
}

// checks if  a struct value is empty
func IsEmpty(data any, collect map[string]string) map[string]string {
	values := reflect.ValueOf(data)
	typesOf := values.Type()
	for i := 0; i < values.NumField(); i++ {
		if values.Field(i).Len() == 0 && typesOf.Field(i).Name != "Errors" {
			collect[typesOf.Field(i).Name] = fmt.Sprintf("%s required", typesOf.Field(i).Name)
		}
	}
	return collect
}

type DocRegister struct {
	Email           string
	Password        string
	Username        string
	Fullname        string
	Contact         string
	ConfirmPassword string
	Departmentname  string
	Errors
}

func (d *DocRegister) validate() (Errors, bool) {
	d.Errors = make(map[string]string)
	err := validateEmail(d.Email)
	if err != nil {
		d.Errors["Email"] = "Please enter a valid email address"
	}
	if d.ConfirmPassword != d.Password {
		d.Errors["Match"] = "Password & ConfirmPassword don't match"
	}
	if len([]rune(d.Password)) < 6 {
		d.Errors["LengthPassword"] = "Password length should be longer then six characters"
	}
	d.Errors = IsEmpty(*d, d.Errors)
	return d.Errors, len(d.Errors) == 0
}

type NurseRegister struct {
	Email           string
	Password        string
	Username        string
	Fullname        string
	ConfirmPassword string
	Errors
}

func (d *NurseRegister) validate() (Errors, bool) {
	d.Errors = make(map[string]string)
	err := validateEmail(d.Email)
	if err != nil {
		d.Errors["Email"] = "Please enter a valid email address"
	}
	if d.ConfirmPassword != d.Password {
		d.Errors["Match"] = "Password & ConfirmPassword don't match"
	}
	if len([]rune(d.Password)) < 6 {
		d.Errors["LengthPassword"] = "Password length should be longer then six characters"
	}
	d.Errors = IsEmpty(*d, d.Errors)
	return d.Errors, len(d.Errors) == 0
}

type Appointment struct {
	Doctorid        string
	Patientid       string
	AppointmentDate string
	Duration        string
	Approval        string
	Errors
}

func (a *Appointment) validate() (Errors, bool) {
	a.Errors = make(map[string]string)
	a.Errors = IsEmpty(*a, a.Errors)
	today := time.Now().Format("2006-01-02T15:04")
	td, _ := time.Parse("2006-01-02T15:04", today)
	appointmentday, _ := time.Parse("2006-01-02T15:04", a.AppointmentDate)
	if appointmentday.Before(td) {
		a.Errors["AppointmentDate Input"] = "You can't travel back to the past,unless you have a time travel machine"
	}
	return a.Errors, len(a.Errors) == 0
}

type PatientAppointment struct {
	AppointmentDate string
	Duration        string
	Errors
}

func (a *PatientAppointment) validate() (Errors, bool) {
	a.Errors = make(map[string]string)
	a.Errors = IsEmpty(*a, a.Errors)
	today := time.Now().Format("2006-01-02T15:04")
	td, _ := time.Parse("2006-01-02T15:04", today)
	appointmentday, _ := time.Parse("2006-01-02T15:04", a.AppointmentDate)
	if appointmentday.Before(td) {
		a.Errors["AppointmentDate Input"] = "You can't travel back to the past,unless you have a time travel machine"
	}
	return a.Errors, len(a.Errors) == 0
}

type Department struct {
	Departmentname string
	Errors
}

func (a *Department) validate() (Errors, bool) {
	a.Errors = make(map[string]string)
	a.Errors = IsEmpty(*a, a.Errors)
	return a.Errors, len(a.Errors) == 0
}

type Schedule struct {
	Doctorid  string
	Starttime string
	Endtime   string
	Active    string
	Errors
}

func (a *Schedule) validate() (Errors, bool) {
	a.Errors = make(map[string]string)
	a.Errors = IsEmpty(*a, a.Errors)
	return a.Errors, len(a.Errors) == 0
}

type Role struct {
	Rolename   string
	Permission string
	Errors
}

func (d *Role) validate() (Errors, bool) {
	d.Errors = make(map[string]string)
	d.Errors = IsEmpty(*d, d.Errors)
	return d.Errors, len(d.Errors) == 0
}

type UpdateRole struct {
	Rolename   string
	Permission []string
	Errors
}

func (d *UpdateRole) validate() (Errors, bool) {
	d.Errors = make(map[string]string)
	d.Errors = IsEmpty(*d, d.Errors)
	return d.Errors, len(d.Errors) == 0
}

type StaffRecords struct {
	Diagnosis    string
	Disease      string
	Prescription string
	Weight       string
	Errors
}

func (d *StaffRecords) validate() (Errors, bool) {
	d.Errors = make(map[string]string)
	d.Errors = IsEmpty(*d, d.Errors)
	return d.Errors, len(d.Errors) == 0
}

type Records struct {
	Patientid   string
	Height      string
	Bp          string
	HeartRate   string
	Temperature string
	Weight      string
	Doctorid    string
	// Additional  string
	Errors
}

func (d *Records) validate() (Errors, bool) {
	d.Errors = make(map[string]string)
	d.Errors = IsEmpty(*d, d.Errors)
	return d.Errors, len(d.Errors) == 0
}

type AdminstrativeUser struct {
	Email           string
	Rolename        string
	Password        string
	ConfirmPassword string
	Errors
}

func (u *AdminstrativeUser) validate() (Errors, bool) {
	u.Errors = make(map[string]string)
	err := validateEmail(u.Email)
	if err != nil {
		u.Errors["Email"] = "Please enter a valid email address"
	}
	if u.ConfirmPassword != u.Password {
		u.Errors["Match"] = "Password & ConfirmPassword don't match"
	}
	u.Errors = IsEmpty(*u, u.Errors)
	return u.Errors, len(u.Errors) == 0
}
