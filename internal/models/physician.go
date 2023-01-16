package models

import "time"

//physician model

type (
	//Physician struct
	Physician struct {
		Physicianid         int       `json:"id,omitempty"`
		Username            string    `json:"username,omitempty"`
		Full_name           string    `json:"fullname,omitempty"`
		Email               string    `json:"email,omitempty"`
		Contact             string    `json:"contact,omitempty"`
		Hashed_password     string    `json:"password,omitempty"`
		Password_changed_at time.Time `json:"password_changed_at,omitempty"`
		Created_at          time.Time `json:"created_at,omitempty"`
		Departmentname      string    `json:"departmentname,omitempty"`
		//verfied string
	}

	ListDoctorsbyDeptarment struct {
		Department string
		Limit      int
		Offset     int
	}
	ListDoctors struct {
		Limit  int
		Offset int
	}

	//Physicianrepository represent the Physician repository contract
	Physicianrepository interface {
		Create(physician Physician) (Physician, error)
		Find(id int) (Physician, error)
		Count() (int, error)
		FindbyEmail(email string) (Physician, error)
		FindAll(ListDoctors) ([]Physician, error)
		FindDoctorsbyDept(ListDoctorsbyDeptarment) ([]Physician, error)
		Delete(id int) error
		Update(physician Physician) (Physician, error)
	}
)
