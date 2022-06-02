package models

import "time"

//patient record model

type Patientrecords struct {
	patienid     int
	doctorid     int
	date         time.Time
	diagnosis    string
	disease      string
	prescription string
	weight       string
}

//Patientrecordsrepository represent the Patientrecords repository contract
type Patientrecordsrepository interface {
	Create(patientrecords Patientrecords) (Patientrecords, error)
	Find(id int) (Patientrecords, error)
	FindAll() ([]Patientrecords, error)
	Delete(id int) error
}
