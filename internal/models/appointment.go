package models

//Appointment table model
import "time"

type (
	Appointment struct {
		Appointmentid   int
		Doctorid        int
		Patientid       int
		Appointmentdate time.Time
	}

	//AppointmentRepository represent the Appointment repository contract
	AppointmentRepository interface {
		Create(appointment Appointment) (Appointment, error)
		Find(id int) (Appointment, error)
		FindAll() ([]Appointment, error)
		Delete(id int) error
		Update(time time.Time, id int) (time.Time, error)
		FindAllByDoctor(id int) ([]Appointment, error)
		FindAllByPatient(id int) ([]Appointment, error)
	}
)
