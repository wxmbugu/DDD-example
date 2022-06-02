package models

import "time"

//Physcian struct
type Physician struct {
	username            string
	full_name           string
	email               string
	hashed_password     string
	password_changed_at time.Time
	created_at          time.Time
	//verfied string
}

//Physicianrepository represent the Physician repository contract
type Physicianrepository interface {
	Create(physician Physician) (Physician, error)
	Find(id int) (Physician, error)
	FindAll() ([]Physician, error)
	Delete(id int) error
}
