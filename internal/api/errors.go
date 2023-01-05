package api

import (
	// "errors"
	"fmt"
	// "log"
	"net/mail"
	"reflect"
)

type validation interface {
	validate() (Errors, bool)
}

// Holds the error value
type Errors map[string]string

type Form struct {
	// check input type content is valid.
	Data validation
	// Holds the error value
	Errors
}

func (f *Form) Validate() bool {
	f.Errors = make(map[string]string)
	if data, ok := validateType(f.Data); !ok {
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
	r.Errors = IsEmpty(*r, r.Errors)
	return r.Errors, len(r.Errors) == 0
}

func validateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	return err
}

// IsEmpty() checks if  a struct value is empty
func IsEmpty(data any, collect map[string]string) map[string]string {
	values := reflect.ValueOf(data)
	typesOf := values.Type()
	for i := 0; i < values.NumField(); i++ {
		if values.Field(i).Len() == 0 && reflect.ValueOf(collect).Kind() != reflect.Map {
			collect[typesOf.Field(i).Name] = fmt.Sprintf("%s required", typesOf.Field(i).Name)
		}
	}
	return collect
}
