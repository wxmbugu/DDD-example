package models

//Appointment table model
import "time"

type (
	Appointment struct {
		Appointmentid   int
		Doctorid        int
		Patientid       int
		Appointmentdate time.Time
		Duration        string
		Approval        bool
	}
	ListAppointments struct {
		Limit  int
		Offset int
	}
	//AppointmentRepository represent the Appointment repository contract
	AppointmentRepository interface {
		Create(appointment Appointment) (Appointment, error)
		Find(id int) (Appointment, error)
		FindAll(ListAppointments) ([]Appointment, error)
		Delete(id int) error
		Update(update Appointment) (Appointment, error)
		FindAllByDoctor(id int) ([]Appointment, error)
		FindAllByPatient(id int) ([]Appointment, error)
	}
)
