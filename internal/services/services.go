package services

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/patienttracker/internal/controllers"
	"github.com/patienttracker/internal/models"
	//"github.com/patienttracker/internal/utils"
)

type Service struct {
	DoctorService      models.Physicianrepository
	AppointmentService models.AppointmentRepository
	ScheduleService    models.Schedulerepositroy
	PatientService     models.PatientRepository
}

func NewService() Service {
	conn, err := sql.Open("postgres", "postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	controllers := controllers.New(conn)
	return Service{
		DoctorService:      controllers.Doctors,
		AppointmentService: controllers.Appointment,
		ScheduleService:    controllers.Schedule,
		PatientService:     controllers.Patient,
	}
}

//This method will help patient book appointment with the doctor and not at odd hours.
//Step 1:Check doctor's Schedule.
//Step 2:Check if the time the patient allocated is within the schedule.
//Step 3:Check if the time the patient allocated has been already occupied by another appointment and gives a range of one hour estimately.
//Step 3.5:If allocated appointment is allocated to anyone it suggests another time slot.
//Step 4:If the book time doesn't fall on any other steps the appointment is booked with the allocated time.
func (service *Service) PatientBookAppointment() (models.Schedule, error) {
	//Start by checking the work schedule of the doctor so as to
	//enable booking for Appointments with the Doctor
	//This will help in making patients not booking apppointments at odd hours
	return models.Schedule{}, nil
}
