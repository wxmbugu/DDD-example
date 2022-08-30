package mock

type Memstore struct {
	PatientMemStore     Patient
	RecordMemStore      PatientRecords
	DoctorMemStore      Doctor
	DepartmentMemStore  Department
	AppointmentMemStore Appointment
	ScheduleMemStore    Schedule
}

func NewMockStore() Memstore {
	return Memstore{
		PatientMemStore:     Patient{},
		RecordMemStore:      PatientRecords{},
		DoctorMemStore:      Doctor{},
		DepartmentMemStore:  Department{},
		AppointmentMemStore: Appointment{},
		ScheduleMemStore:    Schedule{},
	}
}
