package inmem

import (
	"github.com/patienttracker/internal/models"
)

type Memstore struct {
	PatientMemStore     *Patient
	RecordMemStore      *PatientRecords
	DoctorMemStore      *Doctor
	DepartmentMemStore  *Department
	AppointmentMemStore *Appointment
	ScheduleMemStore    *Schedule
}

func NewMockStore() Memstore {
	patientmap := make(map[int]models.Patient)
	doctormap := make(map[int]models.Physician)
	deptmap := make(map[int]models.Department)
	recordmap := make(map[int]models.Patientrecords)
	appointmentmap := make(map[int]models.Appointment)
	schedulemap := make(map[int]models.Schedule)
	return Memstore{
		PatientMemStore: &Patient{
			data: patientmap,
		},
		RecordMemStore: &PatientRecords{
			data: recordmap,
		},
		DoctorMemStore: &Doctor{
			data: doctormap,
		},
		DepartmentMemStore: &Department{
			data: deptmap,
		},
		AppointmentMemStore: &Appointment{
			data: appointmentmap,
		},
		ScheduleMemStore: &Schedule{
			data: schedulemap,
		},
	}
}
