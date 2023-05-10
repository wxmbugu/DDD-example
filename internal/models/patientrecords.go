package models

import "time"

// Take temperature
//   - Weight and Height
//   - Blood Pressure
//   - Heart Rate
//
// patient record model
type (
	Patientrecords struct {
		Recordid int
		Patienid int
		Date     time.Time
		Height   int
		//blood pressure
		Bp          int
		HeartRate   int
		Temperature int
		Weight      string
		Doctorid    int
		Additional  string
		Nurseid     int
	}

	ListPatientRecords struct {
		Limit  int
		Offset int
	}
)

// Patientrecordsrepository represent the Patientrecords repository contract
type Patientrecordsrepository interface {
	Create(patientrecords Patientrecords) (Patientrecords, error)
	Find(id int) (Patientrecords, error)
	FindAll(ListPatientRecords) ([]Patientrecords, error)
	Count() (int, error)
	FindAllByDoctor(id int) ([]Patientrecords, error)
	FindAllByPatient(id int) ([]Patientrecords, error)
	FindAllByNurse(id int) ([]Patientrecords, error)
	Delete(id int) error
	Update(record Patientrecords) (Patientrecords, error)
}
