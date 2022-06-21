package services

import (
	"database/sql"
	"errors"
	"fmt"
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
	schedules, err := service.ScheduleService.FindbyDoctor(doctorid)
	if err != nil {
		log.Fatal(err)

	}
	for i := 0; i < len(schedules); i++ {
		//we check if the time schedule being booked is active
		if schedules[i].Active {
			//layout := "2006-01-02 15:04:05"

			//newLayout := "15:04 EAT"
			//starttime, _ := time.Parse(newLayout, schedules[i].Starttime)
			//fmt.Println(starttime)
			//endtime, _ := time.Parse(newLayout, schedules[i].Endtime)
			//we check if the time being booked is within the working hours of doctors schedule
			fmt.Println("wwwwwwwwwwwwwwwwww", timescheduled.Local())
			fmt.Println(schedules[i].Starttime.Local(), schedules[i].Endtime.Local(), timescheduled.Local())
			fmt.Println(withinTimeFrame(schedules[i].Starttime.Local(), schedules[i].Endtime.Local(), timescheduled.Local()))
			if withinTimeFrame(schedules[i].Starttime.Local(), schedules[i].Endtime.Local(), timescheduled) {
				schedules, err := service.AppointmentService.FindAllByDoctor(doctorid)
				fmt.Println("shhhh", schedules)
				if err != nil {
					log.Fatal(err)
				}
				if len(schedules) <= 0 {
					apt := models.Appointment{
						Doctorid:        doctorid,
						Patientid:       patientid,
						Appointmentdate: timescheduled,
						Duration:        sessionduration.String(),
						Approval:        false,
					}
					appointment, err = service.AppointmentService.Create(apt)
					if err != nil {
						log.Fatal(err)
					}
					return appointment, nil
				}
				fmt.Println("some")
				for i := 0; i < len(schedules); i++ {
					fmt.Println("some2")
					schedule := schedules[i]
					if schedule.Appointmentdate != timescheduled {
						fmt.Println("shiiiiittt", schedule)
						duration, err := time.ParseDuration(schedule.Duration)
						if err != nil {
							log.Fatal(err)
						}
						fmt.Println("duration", schedules)
						endtime := schedule.Appointmentdate.Add(duration)
						fmt.Println("endtime", endtime)
						fmt.Println("starttime", schedule.Appointmentdate)
						fmt.Println("scheduledtime", timescheduled)
						if withinTimeFrame(schedule.Appointmentdate, endtime, timescheduled) && schedule.Approval {
							log.Fatal(ErrTimeSlotAllocated)
							log.Println("weeeee", endtime)
						} else {
							apt := models.Appointment{
								Doctorid:        doctorid,
								Patientid:       patientid,
								Appointmentdate: timescheduled,
								Duration:        sessionduration.String(),
								Approval:        false,
							}
							appointment, err = service.AppointmentService.Create(apt)
							if err != nil {
								log.Fatal(err)
							}
							log.Println("Ok", apt, appointment)
							return appointment, nil
						}
					}
				}

			}
			log.Println(ErrWithinTime)
		} else {

			log.Fatal(ErrInvalidSchedule)
		}
	}
	return appointment, nil
}
