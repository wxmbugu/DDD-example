package controllers

import "database/sql"

type Controllers struct {
	Records     PatientRecords
	Doctors     Physician
	Patient     Patient
	Appointment Appointment
}

func New(conn *sql.DB) Controllers {
	return Controllers{
		Records: PatientRecords{
			db: conn,
		},
		Doctors: Physician{
			db: conn,
		},
		Patient: Patient{
			db: conn,
		},
		Appointment: Appointment{
			db: conn,
		},
	}
}
