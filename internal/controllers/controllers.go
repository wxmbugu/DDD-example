package controllers

import (
	"database/sql"

	"github.com/patienttracker/pkg/logger"
)

type Controllers struct {
	Records     PatientRecords
	Doctors     Physician
	Patient     Patient
	Appointment Appointment
	Schedule    Schedule
	Department  Department
}

func New(conn *sql.DB) Controllers {
	log := logger.New()
	err := conn.Ping()
	if err != nil {
		log.PrintFatal(err)
	}
	log.PrintInfo("Connected to db successfuly")
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
		Schedule: Schedule{
			db: conn,
		},
		Department: Department{
			db: conn,
		},
	}
}
