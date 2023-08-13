package models

import "time"

// patient record model
type (
	Patientrecords struct {
		Recordid int
		Patienid int
		Date     time.Time
		Height   int
		//blood pressure
		Bp          string
		HeartRate   int
		Temperature int
		Weight      string
		Doctorid    int
		Additional  string
		Nurseid     int
	}
)

// Patientrecordsrepository represent the Patientrecords repository contract
type Patientrecordsrepository interface {
	Create(patientrecords Patientrecords) (Patientrecords, error)
	Find(id int) (Patientrecords, error)
	FindAll(Filters) ([]Patientrecords, *Metadata, error)
	FindAllByDoctor(id int) ([]Patientrecords, error)
	FindAllByPatient(id int) ([]Patientrecords, error)
	FindAllByNurse(id int) ([]Patientrecords, error)
	Delete(id int) error
	Update(record Patientrecords) (Patientrecords, error)
}
