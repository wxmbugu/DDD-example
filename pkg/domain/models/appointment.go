package models

//Appointment table model
import "time"

type Appointment struct {
	doctorid        int
	patientid       int
	appointmentdate time.Time
}

//AppointmentRepository represent the Appointment repository contract
type AppointmentRepository interface {
	Create(appointment Appointment) (Appointment, error)
	Find(id int) (Appointment, error)
	FindAll() ([]Appointment, error)
	Delete(id int) error
}
