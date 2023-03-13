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
		About               string
		Verified            bool
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
