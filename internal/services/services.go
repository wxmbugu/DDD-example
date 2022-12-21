package services

import (
	"database/sql"
	"errors"
	_ "github.com/lib/pq"
	"github.com/patienttracker/internal/controllers"
	"github.com/patienttracker/internal/models"
	"log"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	DoctorService        models.Physicianrepository
	AppointmentService   models.AppointmentRepository
	ScheduleService      models.Schedulerepositroy
	PatientService       models.PatientRepository
	DepartmentService    models.Departmentrepository
	PatientRecordService models.Patientrecordsrepository
}

// t wil be the string use to format the appointment dates into 24hr string
const t = "15:00"

var (
	ErrInvalidSchedule   = errors.New("no active shedule found for this doctor")
	ErrTimeSlotAllocated = errors.New("this time slot is already booked")
	ErrNotWithinTime     = errors.New("appointment not within doctors work hours")
	ErrScheduleActive    = errors.New("you should have one schedule active")
	ErrUpdateSchedule    = errors.New("you can only update an active schedule")
	ErrNoUser            = errors.New("no such user")
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

// checks if the time scheduled falls between an appointment already booked with its duration and date
func withinAppointmentTime(start, end, check time.Time) bool {
	if check.Equal(end) && check.After(start) {
		return true
	}
	if check.Equal(start) && check.Before(end) {
		return true
	}
	return check.After(start) && check.Before(end)
}

// This function checks if the time being booked is within the doctors schedule
func withinTimeFrame(start, end, booked float64) bool {
	if booked == start && booked < end {
		return booked > start && booked < end
	}
	if booked == end && booked > start {
		return booked > start && booked < end
	}
	return booked > start && booked < end
}

// this function converts time string into a float64 so something like 14:30
// will be 14.0 then the withintimeframe will check if the time is between the doctors schedule
func formatstring(s string) float64 {
	newstring := strings.Split(s, ":")
	stringtime := strings.Join(newstring, ".")
	time, _ := strconv.ParseFloat(stringtime, 64)
	return time
}

func (service *Service) getallschedules(id int) ([]models.Schedule, error) {
	schedules, err := service.ScheduleService.FindbyDoctor(id)
	return schedules, err
}

func (service *Service) PatientBookAppointment(appointment models.Appointment) (models.Appointment, error) {
	//Start by checking the work schedule of the doctor so as to
	//enable booking for Appointments with the Doctor within doctor's work hours
	schedules, _ := service.getallschedules(appointment.Doctorid)

	schedule, ok := checkschedule(schedules)
	if ok {
		//we check if the time being booked is within the working hours of doctors schedule
		//checks if the appointment boooked is within the doctors schedule
		//if not it errors with ErrWithinTime
		if withinTimeFrame(formatstring(schedule.Starttime), formatstring(schedule.Endtime), formatstring(appointment.Appointmentdate.Format(t))) {
			appointments, _ := service.AppointmentService.FindAllByDoctor(appointment.Doctorid)
			//add appointment after all checks have passed
			appointment, err := service.addappointment(appointments, appointment)
			return appointment, err
		}
		return appointment, ErrNotWithinTime
	}
	return appointment, ErrInvalidSchedule

}
func (service *Service) DoctorBookAppointment(appointment models.Appointment) (models.Appointment, error) {
	//Start by checking the work schedule of the doctor so as to
	//enable booking for Appointments with the Doctor within doctor's work hours
	schedules, err := service.getallschedules(appointment.Doctorid)
	if err != nil {
		log.Fatal(err)
	}
	schedule, ok := checkschedule(schedules)
	if ok {
		//we check if the time being booked is within the working hours of doctors schedule
		//checks if the appointment boooked is within the doctors schedule
		//if not it errors with ErrWithinTime
		//fmt.Println(formatstring(appointment.Appointmentdate.Format(t)), formatstring(schedule.Endtime))
		if withinTimeFrame(formatstring(schedule.Starttime), formatstring(schedule.Endtime), formatstring(appointment.Appointmentdate.Format(t))) {
			appointments, err := service.AppointmentService.FindAllByPatient(appointment.Patientid)
			if err != nil {
				log.Fatal(err)
			}
			//add appointment after all checks have passed
			appointment, err := service.addappointment(appointments, appointment)
			return appointment, err
		}
		return appointment, ErrNotWithinTime
	}
	return appointment, ErrInvalidSchedule
}

// method to add an appointment
func (service *Service) addappointment(appointments []models.Appointment, appointment models.Appointment) (models.Appointment, error) {
	var newappointment models.Appointment
	var err error
	if appointments == nil {
		newappointment, err = service.AppointmentService.Create(appointment)
		if err != nil {
			log.Fatal(err)
		}
		return newappointment, nil
	}
	for _, apntmnt := range appointments {
		duration, err := time.ParseDuration(apntmnt.Duration)
		if err != nil {
			log.Fatal(err)
		}
		endtime := apntmnt.Appointmentdate.Add(duration)

		//fmt.Println("*********", withinAppointmentTime(apntmnt.Appointmentdate, endtime, appointment.Appointmentdate), apntmnt.Approval)
		//checks if there's a booked slot and is approved
		//if there's an appointment within this timeframe it errors with ErrTimeSlotAllocated
		if withinAppointmentTime(apntmnt.Appointmentdate, endtime, appointment.Appointmentdate) && apntmnt.Approval {
			return newappointment, ErrTimeSlotAllocated
		}

		newappointment, err = service.AppointmentService.Create(appointment)
		if err != nil {
			log.Fatal(err)
		}

	}
	return newappointment, nil
}
func (service *Service) UpdateappointmentbyDoctor(doctorid int, appointment models.Appointment) (models.Appointment, error) {
	schedules, err := service.getallschedules(doctorid)
	if err != nil {
		log.Fatal(err)
	}
	appointment, err = service.AppointmentService.Find(appointment.Appointmentid)
	if err != nil {
		return appointment, err
	}
	schedule, ok := checkschedule(schedules)
	if ok {
		if withinTimeFrame(formatstring(schedule.Starttime), formatstring(schedule.Endtime), formatstring(appointment.Appointmentdate.Format(t))) {
			appointments, err := service.AppointmentService.FindAllByDoctor(doctorid)
			if err != nil {
				log.Fatal(err)
			}
			for _, apntmnt := range appointments {
				duration, err := time.ParseDuration(apntmnt.Duration)
				if err != nil {
					log.Fatal(err)
				}
				endtime := apntmnt.Appointmentdate.Add(duration)
				// checks if there's a booked slot and is approved
				// if there's an appointment within this timeframe it errors with ErrTimeSlotAllocated
				if withinAppointmentTime(apntmnt.Appointmentdate, endtime, appointment.Appointmentdate) && apntmnt.Appointmentid != appointment.Appointmentid {
					return appointment, ErrTimeSlotAllocated
				}

				appointment, err = service.AppointmentService.Update(appointment)
				if err != nil {
					log.Fatal(err)
				}
				return appointment, nil
			}
		}
	}
	return appointment, ErrInvalidSchedule
}

func (service *Service) UpdateappointmentbyPatient(patientid int, appointment models.Appointment) (models.Appointment, error) {
	var updatedappointment models.Appointment
	schedules, err := service.getallschedules(appointment.Doctorid)
	if err != nil {
		log.Fatal(err)
	}

	_, ok := checkschedule(schedules)
	if ok {
		appointments, err := service.AppointmentService.FindAllByPatient(patientid)
		if err != nil {
			log.Fatal(err)
		}
		appointment, err = service.AppointmentService.Find(appointment.Appointmentid)
		if err != nil {
			return updatedappointment, err
		}
		for _, apntmnt := range appointments {
			duration, err := time.ParseDuration(apntmnt.Duration)
			if err != nil {
				log.Fatal(err)
			}
			endtime := apntmnt.Appointmentdate.Add(duration)
			//checks if there's a booked slot and is approved
			//if there's an appointment within this timeframe it errors with ErrTimeSlotAllocated
			if withinAppointmentTime(apntmnt.Appointmentdate, endtime, appointment.Appointmentdate) && apntmnt.Appointmentid != appointment.Appointmentid {
				return appointment, ErrTimeSlotAllocated
			}
			updatedappointment, err = service.AppointmentService.Update(appointment)
			if err != nil {
				log.Fatal(err)
			}
		}
		return updatedappointment, nil
	}
	return updatedappointment, ErrInvalidSchedule

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

func (service *Service) UpdateSchedule(schedule models.Schedule) (models.Schedule, error) {
	var newschedule models.Schedule
	activeschedule, err := service.ScheduleService.Find(schedule.Scheduleid)
	if err != nil {
		log.Fatal(err)
	}
	if activeschedule.Active {
		newschedule, err = service.ScheduleService.Update(schedule)
		if err != nil {
			log.Fatal(err)
		}
		return newschedule, nil
	}
	return newschedule, ErrUpdateSchedule
}

func checkschedule(schedules []models.Schedule) (models.Schedule, bool) {
	for _, schedule := range schedules {
		//we check if the time schedule being booked is active
		if schedule.Active {
			return schedule, true
		}
	}
	return models.Schedule{}, false
}
