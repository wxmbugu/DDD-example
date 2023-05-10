package controllers

import (
	"database/sql"
)

type Controllers struct {
	Records     PatientRecords
	Doctors     Physician
	Patient     Patient
	Nurse       Nurse
	Appointment Appointment
	Schedule    Schedule
	Department  Department
	Roles       Roles
	Users       Users
	Permissions Permissions
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
		Schedule: Schedule{
			db: conn,
		},
		Nurse: Nurse{
			db: conn,
		},
		Department: Department{
			db: conn,
		},
		Roles: Roles{
			db: conn,
		},
		Users: Users{
			db: conn,
		},
		Permissions: Permissions{
			db: conn,
		},
	}
}
