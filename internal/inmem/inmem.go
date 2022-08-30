package inmem

import "github.com/patienttracker/internal/services"

type Memstore struct {
	PatientMemStore     *Patient
	RecordMemStore      *PatientRecords
	DoctorMemStore      *Doctor
	DepartmentMemStore  *Department
	AppointmentMemStore *Appointment
	ScheduleMemStore    *Schedule
}

func NewMockStore() Memstore {
	return Memstore{
		PatientMemStore:     &Patient{},
		RecordMemStore:      &PatientRecords{},
		DoctorMemStore:      &Doctor{},
		DepartmentMemStore:  &Department{},
		AppointmentMemStore: &Appointment{},
		ScheduleMemStore:    &Schedule{},
	}
}

func NewServiceMockStore() services.Service {
	memstore := NewMockStore()

	return services.Service{
		DoctorService:        memstore.DoctorMemStore,
		AppointmentService:   memstore.AppointmentMemStore,
		ScheduleService:      memstore.ScheduleMemStore,
		PatientService:       memstore.PatientMemStore,
		DepartmentService:    memstore.DepartmentMemStore,
		PatientRecordService: memstore.RecordMemStore,
	}

}
