package models

import "time"

//patient record model

type Patientrecords struct {
	Recordid     int
	Patienid     int
	Doctorid     int
	Date         time.Time
	Diagnosis    string
	Disease      string
	Prescription string
	Weight       string
}

type PatientrecordsUpd struct {
	Diagnosis    string
	Disease      string
	Prescription string
	Weight       string
}

//Patientrecordsrepository represent the Patientrecords repository contract
type Patientrecordsrepository interface {
	Create(patientrecords Patientrecords) (Patientrecords, error)
	Find(id int) (Patientrecords, error)
	FindAll() ([]Patientrecords, error)
	FindAllByDoctor(id int) ([]Patientrecords, error)
	FindAllByPatient(id int) ([]Patientrecords, error)
	Delete(id int) error
	Update(record PatientrecordsUpd, id int) (Patientrecords, error)
}
