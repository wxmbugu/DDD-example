package models

// Patient model struct
import "time"

type Patient struct {
	username           string
	full_name          string
	email              string
	dob                time.Time
	contact            string
	bloodgroup         string
	hashed_password    string
	password_change_at time.Time
	created_at         time.Time
	//verified           bool
}

//PatientRepository represent the Patient repository contract
type PatientRepository interface {
	Create(patient Patient) (Patient, error)
	Find(id int) (Patient, error)
	FindAll() ([]Patient, error)
	Delete(id int) error
}
