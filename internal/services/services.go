package services

import (
	"database/sql"
	"errors"
	_ "github.com/lib/pq"
	"github.com/patienttracker/internal/controllers"
	"github.com/patienttracker/internal/models"
	"strconv"
	"strings"
	"time"
)

// role based access contriol to adminster the service.
type Rbac struct {
	RolesService       models.RolesRepository
	UsersService       models.UsersRepository
	PermissionsService models.PermissionsRepository
}

type Service struct {
	DoctorService        models.Physicianrepository
	AppointmentService   models.AppointmentRepository
	ScheduleService      models.Schedulerepositroy
	PatientService       models.PatientRepository
	DepartmentService    models.Departmentrepository
	PatientRecordService models.Patientrecordsrepository
	RbacService          Rbac
}

// t wil be the string use to format the appointment dates into 24hr string
const t = "15:00"

var (
	ErrInvalidSchedule    = errors.New("no active shedule found for this doctor")
	ErrTimeSlotAllocated  = errors.New("this time slot is already booked")
	ErrNotWithinTime      = errors.New("appointment not within doctors work hours")
	ErrScheduleActive     = errors.New("you should have one schedule active")
	ErrUpdateSchedule     = errors.New("you can only have one active schedule")
	ErrNoUser             = errors.New("no such user")
	ErrInvalidPermissions = errors.New("no such permission available")
	ErrNotAuthorized      = errors.New("you don't have the required permissions to execute this task")
	ErrForbidden          = errors.New("Forbidden")
)

func NewService(conn *sql.DB) Service {
	controllers := controllers.New(conn)
	return Service{
		DoctorService:        controllers.Doctors,
		AppointmentService:   &controllers.Appointment,
		ScheduleService:      controllers.Schedule,
		PatientService:       controllers.Patient,
		DepartmentService:    controllers.Department,
		PatientRecordService: controllers.Records,
		RbacService: Rbac{
			RolesService:       &controllers.Roles,
			UsersService:       &controllers.Users,
			PermissionsService: &controllers.Permissions,
		},
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

func (service *Service) CreateAdmin(email string, password string) (models.Users, error) {
	hashedpass, err := HashPassword(password)
	if err != nil {
		return models.Users{}, err
	}
	role, err := service.RbacService.RolesService.Create(models.Roles{
		Role: "admin",
	})
	if err != nil {
		return models.Users{}, err
	}

	admin, err := service.RbacService.UsersService.Create(models.Users{
		Email:    email,
		Password: hashedpass,
		Roleid:   role.Roleid,
	})
	if err != nil {
		return models.Users{}, err
	}
	if _, err = service.RbacService.PermissionsService.Create(models.Permissions{
		Roleid:     role.Roleid,
		Permission: Admin.toString(),
	}); err != nil {
		return models.Users{}, err
	}
	return admin, err
}

func (service *Service) UpdateRolePermissions(permissions []string, roleid int) error {
	var oldpermissions []string
	permissionfrequency := make(map[string]int)
	availableperimissions, err := service.RbacService.PermissionsService.FindbyRoleId(roleid)
	if err != nil {
		return err
	}
	for _, perm := range availableperimissions {
		oldpermissions = append(oldpermissions, perm.Permission)
	}
	concatpermissions := append(oldpermissions, permissions...)
	for _, perm := range concatpermissions {
		permissionfrequency[perm] += 1
	}
	for _, perm := range oldpermissions {
		if permissionfrequency[perm] == 1 {
			permissionfrequency[perm] -= 1
		}
	}
	for permission := range permissionfrequency {
		switch {
		case permissionfrequency[permission] == 0:
			var perm_ids []int
			for _, perm := range availableperimissions {
				if perm.Permission == permission {
					perm_ids = append(perm_ids, perm.Permissionid)
				}
			}
			for _, id := range perm_ids {
				service.RbacService.PermissionsService.Delete(id)
			}
		case permissionfrequency[permission] == 1:
			_, err := service.RbacService.PermissionsService.Create(models.Permissions{
				Permission: permission,
				Roleid:     roleid,
			})
			if err != nil {
				return err
			}
		case permissionfrequency[permission] == 2:
			// Do nothing because the permissions remain the same
		default:
			break
		}
		delete(permissionfrequency, permission)
	}
	return nil
}

func (service *Service) PatientBookAppointment(appointment models.Appointment) (models.Appointment, error) {
	//Start by checking the work schedule of the doctor so as to
	//enable booking for Appointments with the Doctor within doctor's work hours
	schedules, _ := service.getallschedules(appointment.Doctorid)

	if schedule, ok := checkschedule(schedules); ok {
		//we check if the time being booked is within the working hours of doctors schedule
		//checks if the appointment boooked is within the doctors schedule
		//if not it errors with ErrWithinTime
		if withinTimeFrame(formatstring(schedule.Starttime), formatstring(schedule.Endtime), formatstring(appointment.Appointmentdate.Format(t))) {
			appointments, _ := service.AppointmentService.FindAllByPatient(appointment.Patientid)
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
		return appointment, err
	}
	if schedule, ok := checkschedule(schedules); ok {
		//we check if the time being booked is within the working hours of doctors schedule
		//checks if the appointment boooked is within the doctors schedule
		//if not it errors with ErrWithinTime

		if withinTimeFrame(formatstring(schedule.Starttime), formatstring(schedule.Endtime), formatstring(appointment.Appointmentdate.Format(t))) {
			appointments, err := service.AppointmentService.FindAllByDoctor(appointment.Doctorid)
			if err != nil {
				return appointment, err
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
			return appointment, err
		}
		return newappointment, nil
	}
	if err := checkbooked(appointments, appointment); err != nil {
		return newappointment, err
	}
	newappointment, err = service.AppointmentService.Create(appointment)
	if err != nil {
		return newappointment, err
	}
	return newappointment, nil
}

func checkbooked(appointments []models.Appointment, appointment models.Appointment) error {
	for _, apntmnt := range appointments {
		duration, err := time.ParseDuration(apntmnt.Duration)
		if err != nil {
			return err
		}
		endtime := apntmnt.Appointmentdate.Add(duration)
		// checks if there's a booked slot and is approved
		// if there's an appointment within this timeframe it errors with ErrTimeSlotAllocate
		if withinAppointmentTime(apntmnt.Appointmentdate, endtime, appointment.Appointmentdate) && apntmnt.Approval && appointment.Appointmentid != apntmnt.Appointmentid {
			return ErrTimeSlotAllocated
		}
	}
	return nil
}

func (service *Service) UpdateappointmentbyDoctor(doctorid int, appointment models.Appointment) (models.Appointment, error) {
	var updatedappointment models.Appointment
	schedules, err := service.getallschedules(doctorid)
	if err != nil {
		return updatedappointment, err
	}
	if schedule, ok := checkschedule(schedules); ok {
		if withinTimeFrame(formatstring(schedule.Starttime), formatstring(schedule.Endtime), formatstring(appointment.Appointmentdate.Format(t))) {
			appointments, err := service.AppointmentService.FindAllByDoctor(doctorid)
			if err != nil {
				return updatedappointment, err
			}
			if err := checkbooked(appointments, appointment); err != nil {
				return updatedappointment, err
			}
			updatedappointment, err = service.AppointmentService.Update(appointment)
			if err != nil {
				return updatedappointment, err
			}
			return updatedappointment, nil
		}
	}
	return updatedappointment, ErrInvalidSchedule
}

func (service *Service) UpdateappointmentbyPatient(patientid int, appointment models.Appointment) (models.Appointment, error) {
	var updatedappointment models.Appointment
	schedules, err := service.getallschedules(appointment.Doctorid)
	if err != nil {
		return updatedappointment, err
	}

	if _, ok := checkschedule(schedules); ok {
		appointments, err := service.AppointmentService.FindAllByPatient(patientid)
		if err != nil {
			return updatedappointment, err
		}
		if err := checkbooked(appointments, appointment); err != nil {
			return updatedappointment, err
		}
		updatedappointment, err = service.AppointmentService.Update(appointment)
		if err != nil {
			return updatedappointment, err
		}
		return updatedappointment, nil
	}
	return updatedappointment, ErrInvalidSchedule

}
func (service *Service) MakeSchedule(schedule models.Schedule) (models.Schedule, error) {
	schedules, err := service.ScheduleService.FindbyDoctor(schedule.Doctorid)
	if err != nil {
		return models.Schedule{}, err
	}
	for i := 0; i < len(schedules); i++ {
		//checks if there's an active schedule already
		if schedules[i].Active && schedule.Active {
			return schedule, ErrScheduleActive
		}
	}
	schedule, err = service.ScheduleService.Create(schedule)
	if err != nil {
		return schedule, err
	}
	return schedule, nil
}

func (service *Service) UpdateSchedule(schedule models.Schedule) (models.Schedule, error) {
	var newschedule models.Schedule
	schedules, err := service.ScheduleService.FindbyDoctor(schedule.Doctorid)
	if err != nil {
		return newschedule, err
	}
	var active_schedule []models.Schedule
	for _, schedule := range schedules {
		//we check if the time schedule being booked is active
		if schedule.Active {
			active_schedule = append(active_schedule, schedule)
		}
	}
	if _, err := service.ScheduleService.Find(schedule.Scheduleid); err == nil {
		if len(active_schedule) <= 1 {
			if newschedule, err = service.ScheduleService.Update(schedule); err != nil {
				return newschedule, err
			}
			return newschedule, nil
		}
		return newschedule, ErrUpdateSchedule
	}
	return newschedule, errors.New("no schedule found")
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

func (s *Service) GetAllPermissionsofUser(userid int) ([]models.Permissions, error) {
	user, err := s.RbacService.UsersService.Find(userid)
	if err != nil {
		return nil, errors.New("No such role")
	}
	permissione, err := s.RbacService.PermissionsService.FindbyRoleId(user.Roleid)
	return permissione, nil
}
