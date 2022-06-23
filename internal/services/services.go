package services

import (
	"database/sql"
	"errors"

	//"fmt"
	"log"
	"strconv"
	"strings"
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
	ErrScheduleActive    = errors.New("you should have one schedule active")
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
//func withinTimeFrame(start, end, check time.Time) bool {
//if start.Before(end) {
//return !check.Before(start) && !check.After(end)
//}
//if start.Equal(end) {
//return check.Equal(start)
//}
//return !start.After(check) || !end.Before(check)
//}

//THis function checks if the time being booked is within the doctors schedule
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

//This method will help patient book appointment with the doctor and not at odd hours.
//Step 1:Check doctor's Schedule.
//Step 2:Check if the time the patient allocated is within the schedule.
//Step 3:Check if the time the patient allocated has been already occupied by another appointment and gives a range of one hour estimately.
//Step 3.5:If allocated appointment is allocated to anyone it suggests another time slot.
//Step 4:If the book time doesn't fall on any other steps the appointment is booked with the allocated time.
func (service *Service) BookAppointment(doctorid int, patientid int, timescheduled time.Time, sessionduration time.Duration, approval bool) (models.Appointment, error) {
	//Start by checking the work schedule of the doctor so as to
	//enable booking for Appointments with the Doctor
	//This will help in making apppointments at within the doctors work hours
	var appointment models.Appointment
	//t wil be the string use to format the appointment dates into 24hr string
	t := "15:00"
	apt := models.Appointment{
		Doctorid:        doctorid,
		Patientid:       patientid,
		Appointmentdate: timescheduled.Local(),
		Duration:        sessionduration.String(),
		Approval:        approval,
	}
	schedules, err := service.ScheduleService.FindbyDoctor(doctorid)
	if err != nil {
		log.Fatal(err)
	}
	for _, schedule := range schedules {
		//we check if the time schedule being booked is active
		if schedule.Active {
			//we check if the time being booked is within the working hours of doctors schedule
			//checks if the appointment boooked is within the doctors schedule
			//if not it errors with ErrWithinTime
			if withinTimeFrame(formatstring(schedule.Starttime), formatstring(schedule.Endtime), formatstring(timescheduled.Format(t))) {
				//checks all doctors appointment
				apppointments, err := service.AppointmentService.FindAllByDoctor(doctorid)
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
				for _, appointment := range apppointments {
					//checks if the scheduled appointmnt is same as the appointment which has already been appproved by the doctor
					if appointment.Appointmentdate != timescheduled && !appointment.Approval {
						duration, err := time.ParseDuration(appointment.Duration)
						if err != nil {
							log.Fatal(err)
						}
						endtime := appointment.Appointmentdate.Add(duration)
						//checks if there's a booked slot and is approved
						//if there's an appointment within this timeframe it errors with ErrTimeSlotAllocated
						if withinTimeFrame(formatstring(appointment.Appointmentdate.Format(t)), formatstring(endtime.Format(t)), formatstring(timescheduled.Format(t))) && appointment.Approval {
							return appointment, ErrTimeSlotAllocated
						}
						appointment, err = service.AppointmentService.Create(apt)
						if err != nil {
							log.Fatal(err)
						}
						return appointment, nil
					} else {

						return appointment, ErrTimeSlotAllocated
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

//MakeSchedule is a method for doctors to make a schedule
//The method should check if there's a schedule already active

func (service *Service) MakeSchedule(doctorid int, starttime, endtime string, active bool) (models.Schedule, error) {
	var schedule models.Schedule
	newschedule := models.Schedule{
		Doctorid:  doctorid,
		Starttime: starttime,
		Endtime:   endtime,
		Active:    active,
	}
	schedules, err := service.ScheduleService.FindbyDoctor(doctorid)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(schedules); i++ {
		//checks if there's an active schedule already
		if schedules[i].Active {
			return schedule, ErrScheduleActive
		}
	}
	schedule, err = service.ScheduleService.Create(newschedule)
	if err != nil {
		log.Fatal(err)
	}
	return schedule, nil
}
