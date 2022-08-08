package services

import (
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
	"github.com/stretchr/testify/require"
	//"go.mongodb.org/mongo-driver/mongo/description"
)

var services Service

func TestMain(m *testing.M) {
	conn, err := sql.Open("postgres", "postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	services = NewService(conn)
	os.Exit(m.Run())
}

func RandPatient() models.Patient {
	username := utils.RandUsername(6)
	contact := utils.RandContact(10)
	email := utils.RandEmail(5)
	fname := utils.Randfullname(4)
	date := utils.Randate()
	patient, _ := services.PatientService.Create(models.Patient{
		Username:        username,
		Full_name:       fname,
		Email:           email,
		Dob:             date,
		Contact:         contact,
		Bloodgroup:      utils.RandString(1),
		Hashed_password: utils.RandString(8),
	})
	return patient
}

func RandDoctor() models.Physician {
	username := utils.RandUsername(6)
	email := utils.RandEmail(5)
	fname := utils.Randfullname(4)
	deptname, _ := services.DepartmentService.Create(utils.RandString(6))
	//date := utils.Randate()
	doctor, _ := services.DoctorService.Create(models.Physician{
		Username:        username,
		Full_name:       fname,
		Email:           email,
		Hashed_password: utils.RandString(8),
		Contact:         utils.RandContact(10),
		Departmentname:  deptname.Departmentname,
	})
	return doctor
}

func CreateAppointment(patientid int, doctorid int) models.Appointment {
	//time := utils.Randate()
	appointment, _ := services.AppointmentService.Create(models.Appointment{
		Patientid:       patientid,
		Doctorid:        doctorid,
		Appointmentdate: time.Now().Local(),
		Duration:        "1h",
		Approval:        false,
	})
	return appointment
}

func CreateSchedule(id int) models.Schedule {
	schedule, _ := services.ScheduleService.Create(models.Schedule{
		Doctorid:  id,
		Starttime: "07:00",
		Endtime:   "20:00",
		Active:    true,
	})
	return schedule
}

func TestDoctorBookAppointmentService(t *testing.T) {
	doctor := RandDoctor()
	patient := RandPatient()
	CreateSchedule(doctor.Physicianid)
	duration, _ := time.ParseDuration("1h")
	newappointment := models.Appointment{
		Doctorid:        doctor.Physicianid,
		Patientid:       patient.Patientid,
		Appointmentdate: time.Now(),
		Duration:        duration.String(),
		Approval:        true,
	}
	appointment, err := services.DoctorBookAppointment(newappointment)
	require.NoError(t, err)
	require.NotNil(t, appointment)
	anotherappointment, err := services.DoctorBookAppointment(appointment)
	//fmt.Println("error>>>>", ErrNotWithinTime, err.Error())
	require.EqualError(t, ErrTimeSlotAllocated, err.Error())
	require.Empty(t, anotherappointment)
	appupdate := models.Appointment{
		Appointmentid:   appointment.Appointmentid,
		Appointmentdate: time.Now(),
		Duration:        duration.String(),
		Approval:        true,
	}

	updatedappointment, err := services.UpdateappointmentbyDoctor(appointment.Patientid, appupdate)
	require.NoError(t, err)
	require.NotEmpty(t, updatedappointment)
}
func TestPatientBookAppointmentService(t *testing.T) {
	doctor := RandDoctor()
	patient := RandPatient()
	CreateSchedule(doctor.Physicianid)
	duration, _ := time.ParseDuration("1h")
	newappointment := models.Appointment{
		Doctorid:        doctor.Physicianid,
		Patientid:       patient.Patientid,
		Appointmentdate: time.Now(),
		Duration:        duration.String(),
		Approval:        true,
	}
	appointment, err := services.PatientBookAppointment(newappointment)
	require.NoError(t, err)
	require.NotNil(t, appointment)
	anotherappointment, err := services.PatientBookAppointment(appointment)
	require.EqualError(t, err, ErrTimeSlotAllocated.Error())
	require.Empty(t, anotherappointment)
	appupdate := models.Appointment{
		Appointmentid:   appointment.Appointmentid,
		Appointmentdate: time.Now(),
		Duration:        duration.String(),
		Approval:        false,
	}

	updatedappointment, err := services.UpdateappointmentbyPatient(appointment.Doctorid, appupdate)
	require.NoError(t, err)
	require.NotEmpty(t, updatedappointment)
}

func TestCreateSchedule(t *testing.T) {
	doctor := RandDoctor()
	newschedule := models.Schedule{
		Doctorid:  doctor.Physicianid,
		Starttime: "08:00",
		Endtime:   "20:00",
		Active:    true,
	}
	schedule, err := services.MakeSchedule(newschedule)
	require.NoError(t, err)
	require.NotEmpty(t, schedule)
}

func TestUpdateSchedule(t *testing.T) {
	doctor := RandDoctor()
	newschedule := models.Schedule{
		Doctorid:  doctor.Physicianid,
		Starttime: "08:00",
		Endtime:   "15:00",
		Active:    true,
	}
	schedule, err := services.MakeSchedule(newschedule)
	require.NoError(t, err)
	require.NotEmpty(t, schedule)
	updateschedule := models.Schedule{
		Scheduleid: schedule.Scheduleid,
		Starttime:  "08:00",
		Endtime:    "15:00",
		Active:     true,
	}
	scheduleupdate, err := services.UpdateSchedule(updateschedule)
	require.NoError(t, err)
	require.NotEmpty(t, scheduleupdate)
}
