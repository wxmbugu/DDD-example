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
		//verfied string
	}

	//update Physcian struct
	UpdatePhysician struct {
		Username            string
		Full_name           string
		Email               string
		Contact             string
		Hashed_password     string
		Password_changed_at time.Time
		Departmentname      string
	}

	//Physicianrepository represent the Physician repository contract
	Physicianrepository interface {
		Create(physician Physician) (Physician, error)
		Find(id int) (Physician, error)
		FindAll() ([]Physician, error)
		FindDoctorsbyDept(deptname string) ([]Physician, error)
		Delete(id int) error
		Update(physician UpdatePhysician, id int) (Physician, error)
	}
)
