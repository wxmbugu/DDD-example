package services

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
	"github.com/stretchr/testify/require"
)

var services Service

func TestMain(m *testing.M) {

	services = NewService()
	os.Exit(m.Run())
}

func RandPatient() models.Patient {
	username := utils.RandUsername(6)
	contact := utils.RandContact(10)
	email := utils.RandEmail(5)
	fname := utils.Randfullname(4)
	date := utils.Randate()
	return models.Patient{
		Username:        username,
		Full_name:       fname,
		Email:           email,
		Dob:             date,
		Contact:         contact,
		Bloodgroup:      utils.RandString(1),
		Hashed_password: utils.RandString(8),
	}
}

func RandDoctor() models.Physician {
	username := utils.RandUsername(6)
	email := utils.RandEmail(5)
	fname := utils.Randfullname(4)
	deptname, _ := services.DepartmentService.Create(utils.RandString(6))
	//date := utils.Randate()
	return models.Physician{
		Username:        username,
		Full_name:       fname,
		Email:           email,
		Hashed_password: utils.RandString(8),
		Contact:         utils.RandContact(10),
		Departmentname:  deptname.Departmentname,
	}
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
	ok := time.Date(2022, time.December, 2, 12, 32, 12, 126, time.Local)
	//starttime := "09:00"
	//endtime := "15:00"
	return models.Schedule{
		Doctorid:  id,
		Type:      "monthly",
		Starttime: time.Now().Local(),
		Endtime:   ok,
		Active:    true,
	}
}
func TestSomeService(t *testing.T) {
	doctor := RandDoctor()
	physcian, err := services.DoctorService.Create(doctor)
	require.NoError(t, err)
	require.NotEmpty(t, physcian)
	require.Equal(t, doctor.Email, physcian.Email)
	patient := RandPatient()
	patient1, err := services.PatientService.Create(patient)
	require.NoError(t, err)
	require.NotEmpty(t, patient1)
	require.Equal(t, patient.Email, patient1.Email)
	schedule := CreateSchedule(physcian.Physicianid)
	schedule1, err := services.ScheduleService.Create(schedule)
	fmt.Println("ssss", schedule1)
	require.NoError(t, err)
	require.NotEmpty(t, schedule1)
	duration, _ := time.ParseDuration("1h")
	app := CreateAppointment(patient1.Patientid, physcian.Physicianid)
	app, err = services.AppointmentService.Create(app)
	require.NoError(t, err)
	require.NotEmpty(t, app)
	tme, _ := time.ParseDuration("5h")
	appointment, err := services.PatientBookAppointment(physcian.Physicianid, patient1.Patientid, time.Now().Local().Add(tme), duration)
	require.NoError(t, err)
	fmt.Println(appointment)
	require.NotEmpty(t, appointment)
}
