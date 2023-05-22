package models

import "time"

//physician model

type (
	//Nurse struct
	Nurse struct {
		Id                  int
		Username            string
		Full_name           string
		Email               string
		Hashed_password     string
		Password_changed_at time.Time
		Created_at          time.Time
	}

	//Physicianrepository represent the Physician repository contract
	Nurserepository interface {
		Create(Nurse) (Nurse, error)
		Find(id int) (Nurse, error)
		FindbyEmail(email string) (Nurse, error)
		FindAll(Filters) ([]Nurse, *Metadata, error)
		Filter(string, Filters) ([]*Nurse, *Metadata, error)
		Delete(id int) error
		Update(Nurse) (Nurse, error)
	}
)
