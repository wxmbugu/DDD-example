package services

import (
	"database/sql"
	"errors"

	//"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	//"github.com/golang/protobuf/ptypes/duration"

	_ "github.com/lib/pq"
	"github.com/patienttracker/internal/controllers"
	"github.com/patienttracker/internal/models"
	//"google.golang.org/genproto/googleapis/cloud/scheduler/v1"
	//"github.com/patienttracker/internal/utils"
)

type Service struct {
	DoctorService        models.Physicianrepository
	AppointmentService   models.AppointmentRepository
	ScheduleService      models.Schedulerepositroy
	PatientService       models.PatientRepository
	DepartmentService    models.Departmentrepository
	PatientRecordService models.Patientrecordsrepository
}

//t wil be the string use to format the appointment dates into 24hr string
const t = "15:00"

var (
	ErrInvalidSchedule   = errors.New("invalid schedule it's not active")
	ErrTimeSlotAllocated = errors.New("this time slot is already booked")
	ErrNotWithinTime     = errors.New("appointment not within doctors work hours")
	ErrScheduleActive    = errors.New("you should have one schedule active")
)

func NewService(conn *sql.DB) Service {
	controllers := controllers.New(conn)
	return Service{
		DoctorService:        controllers.Doctors,
		AppointmentService:   controllers.Appointment,
		ScheduleService:      controllers.Schedule,
		PatientService:       controllers.Patient,
		DepartmentService:    controllers.Department,
		PatientRecordService: controllers.Records,
	}
}

//This function checks if the time being booked is within the doctors schedule
//also checks if the time scheduled falles between an appointment already booked with its duration
//rereference//->https://go.dev/play/p/79tgQLCd9f
//https://stackoverflow.com/questions/20924303/how-to-do-date-time-comparison
//func withinTimeFrame(start, end, check time.Time) bool {
//if start.Before(end) {
//return !check.Before(start) && !check.After(end)
//}
//if start.Equal(end) {
//return check.Equal(start)
//}
//return !start.After(check) || !end.Before(check)
//}

//This function checks if the time being booked is within the doctors schedule
//also checks if the time scheduled falles between an appointment already booked with its duration
func withinTimeFrame(start, end, booked float64) bool {
	return booked > start && booked < end
}

//this function converts time string into a float64 so something like 14:56
//will be 14.56 then the withintimeframe will check if the time is between the doctors schedule
func formatstring(s string) float64 {
	newstring := strings.Split(s, ":")
	stringtime := strings.Join(newstring, ".")
	time, _ := strconv.ParseFloat(stringtime, 64)
	return time
}

func (service *Service) BookAppointment(appointment models.Appointment) (models.Appointment, error) {
	//Start by checking the work schedule of the doctor so as to
	//enable booking for Appointments with the Doctor within doctor's work hours
	schedules, err := service.ScheduleService.FindbyDoctor(appointment.Doctorid)
	if err != nil {
		log.Fatal(err)
	}
	for _, schedule := range schedules {
		//we check if the time schedule being booked is active
		if schedule.Active {
			//we check if the time being booked is within the working hours of doctors schedule
			//checks if the appointment boooked is within the doctors schedule
			//if not it errors with ErrWithinTime
			if withinTimeFrame(formatstring(schedule.Starttime), formatstring(schedule.Endtime), formatstring(appointment.Appointmentdate.Format(t))) {
				apppointments, err := service.AppointmentService.FindAllByDoctor(appointment.Doctorid)
				if err != nil {
					log.Fatal(err)
				}
				//add appointment after all checks have passed
				appointment, err := service.addappointment(apppointments, appointment)
				if err != nil {
					log.Fatal(err)
				}
				return appointment, nil
			}
			return appointment, ErrNotWithinTime
		} else {
			return appointment, ErrInvalidSchedule
		}
	}
	return appointment, nil
}

//method to add an appointment
func (service *Service) addappointment(appointments []models.Appointment, appointment models.Appointment) (models.Appointment, error) {
	if appointments == nil {
		appointment, err := service.AppointmentService.Create(appointment)
		if err != nil {
			log.Fatal(err)
		}
		return appointment, nil
	}
	for _, apntmnt := range appointments {
		duration, err := time.ParseDuration(apntmnt.Duration)
		if err != nil {
			log.Fatal(err)
		}
		endtime := apntmnt.Appointmentdate.Add(duration)
		//checks if there's a booked slot and is approved
		//if there's an appointment within this timeframe it errors with ErrTimeSlotAllocated
		if withinTimeFrame(formatstring(apntmnt.Appointmentdate.Format(t)), formatstring(endtime.Format(t)), formatstring(appointment.Appointmentdate.Format(t))) && apntmnt.Approval {
			return appointment, ErrTimeSlotAllocated
		}
		appointment, err = service.AppointmentService.Create(appointment)
		if err != nil {
			log.Fatal(err)
		}

	}
	return appointment, nil
}

func (service *Service) MakeSchedule(schedule models.Schedule) (models.Schedule, error) {
	schedules, err := service.ScheduleService.FindbyDoctor(schedule.Doctorid)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(schedules); i++ {
		//checks if there's an active schedule already
		if schedules[i].Active {
			return schedule, ErrScheduleActive
		}
	}
	schedule, err = service.ScheduleService.Create(schedule)
	if err != nil {
		log.Fatal(err)
	}
	return schedule, nil
}
