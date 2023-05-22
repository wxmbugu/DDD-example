package models

import "time"

//physician model

type (
	//Physician struct
	Physician struct {
		Physicianid         int
		Username            string
		Full_name           string
		Email               string
		Contact             string
		Hashed_password     string
		Password_changed_at time.Time
		Created_at          time.Time
		Departmentname      string
		Avatar              string
		About               string
		Verified            bool
	}

	ListDoctors struct {
		Limit        int
		Offset       int
		Sort         string
		SortSafeList []string
	}

	//Physicianrepository represent the Physician repository contract
	Physicianrepository interface {
		Create(physician Physician) (Physician, error)
		Find(id int) (Physician, error)
		FindbyEmail(email string) (Physician, error)
		FindAll(Filters) ([]Physician, *Metadata, error)
		Filter(string, string, Filters) ([]*Physician, *Metadata, error)
		FindDoctorsbyDept(string, Filters) ([]Physician, *Metadata, error)
		Delete(id int) error
		Update(physician Physician) (Physician, error)
	}
)
