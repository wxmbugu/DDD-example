package services

import (
	"database/sql"
	"errors"
	//"fmt"
	"log"
	"time"

	//	"github.com/golang/protobuf/ptypes/duration"

	_ "github.com/lib/pq"
	"github.com/patienttracker/internal/controllers"
	"github.com/patienttracker/internal/models"
	//"google.golang.org/genproto/googleapis/cloud/scheduler/v1"
	//"github.com/patienttracker/internal/utils"
)

type Service struct {
	DoctorService      models.Physicianrepository
	AppointmentService models.AppointmentRepository
	ScheduleService    models.Schedulerepositroy
	PatientService     models.PatientRepository
	DepartmentService  models.Departmentrepository
}

var (
	ErrInvalidSchedule   = errors.New("invalid schedule it's not active")
	ErrTimeSlotAllocated = errors.New("this time slot is already booked")
	ErrWithinTime        = errors.New("within time duration not within doctors work hours")
)

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
		DepartmentService:  controllers.Department,
	}
}

//THis function checks if the time being booked is within the doctors schedule
//also checks if the time scheduled falles between an appointment already booked with its duration
//rereference//->https://go.dev/play/p/79tgQLCd9f
//https://stackoverflow.com/questions/20924303/how-to-do-date-time-comparison
func withinTimeFrame(start, end, check time.Time) bool {
	if start.Before(end) {
		return !check.Before(start) && !check.After(end)
	}
	if start.Equal(end) {
		return check.Equal(start)
	}
	return !start.After(check) || !end.Before(check)
}

//This method will help patient book appointment with the doctor and not at odd hours.
//Step 1:Check doctor's Schedule.
//Step 2:Check if the time the patient allocated is within the schedule.
//Step 3:Check if the time the patient allocated has been already occupied by another appointment and gives a range of one hour estimately.
//Step 3.5:If allocated appointment is allocated to anyone it suggests another time slot.
//Step 4:If the book time doesn't fall on any other steps the appointment is booked with the allocated time.
func (service *Service) PatientBookAppointment(doctorid int, patientid int, timescheduled time.Time, sessionduration time.Duration) (models.Appointment, error) {
	//Start by checking the work schedule of the doctor so as to
	//enable booking for Appointments with the Doctor
	//This will help in making patients not booking apppointments at odd hours
	var appointment models.Appointment
	apt := models.Appointment{
		Doctorid:        doctorid,
		Patientid:       patientid,
		Appointmentdate: timescheduled,
		Duration:        sessionduration.String(),
		Approval:        false,
	}
	schedules, err := service.ScheduleService.FindbyDoctor(doctorid)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(schedules); i++ {
		//we check if the time schedule being booked is active
		if schedules[i].Active {
			//we check if the time being booked is within the working hours of doctors schedule
			if withinTimeFrame(schedules[i].Starttime.Local(), schedules[i].Endtime.Local(), timescheduled) {
				schedules, err := service.AppointmentService.FindAllByDoctor(doctorid)
				if err != nil {
					log.Fatal(err)
				}
				if schedules == nil {

					appointment, err = service.AppointmentService.Create(apt)
					if err != nil {
						log.Fatal(err)
					}
					return appointment, nil
				}
				for i := 0; i < len(schedules); i++ {
					schedule := schedules[i]
					if schedule.Appointmentdate != timescheduled {
						duration, err := time.ParseDuration(schedule.Duration)
						if err != nil {
							log.Fatal(err)
						}
						endtime := schedule.Appointmentdate.Add(duration)
						if withinTimeFrame(schedule.Appointmentdate, endtime, timescheduled) && schedule.Approval {
							return appointment, ErrTimeSlotAllocated
						} else {
							appointment, err = service.AppointmentService.Create(apt)
							if err != nil {
								log.Fatal(err)
							}
							return appointment, nil
						}
					}
				}
			}
			return appointment, ErrWithinTime
		} else {
			return appointment, ErrInvalidSchedule
		}
	}
	return appointment, nil
}
