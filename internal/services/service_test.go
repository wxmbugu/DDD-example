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
		Patientid:       1,
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
	deptname, _ := services.DepartmentService.Create(models.Department{Departmentname: utils.RandString(6)})
	doctor, _ := services.DoctorService.Create(models.Physician{
		Physicianid:     1,
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
	appointment, _ := services.AppointmentService.Create(models.Appointment{
		Patientid:       patientid,
		Doctorid:        doctorid,
		Appointmentdate: time.Now(),
		Duration:        "1h",
		Approval:        false,
	})
	return appointment
}

func CreateSchedule(id int) models.Schedule {
	schedule, _ := services.ScheduleService.Create(models.Schedule{
		Scheduleid: 1,
		Doctorid:   id,
		Starttime:  "04:00",
		Endtime:    "23:00",
		Active:     true,
	})

	return schedule
}

func newappointment() models.Appointment {
	doctor := RandDoctor()
	patient := RandPatient()
	CreateSchedule(doctor.Physicianid)
	duration, _ := time.ParseDuration("1h")
	return models.Appointment{
		Doctorid:        doctor.Physicianid,
		Patientid:       patient.Patientid,
		Appointmentdate: time.Now().UTC(),
		Duration:        duration.String(),
		Approval:        true,
	}
}

func Doctor_with_no_schedule_appointment() models.Appointment {
	doctor := RandDoctor()
	patient := RandPatient()
	duration, _ := time.ParseDuration("1h")
	return models.Appointment{
		Doctorid:        doctor.Physicianid,
		Patientid:       patient.Patientid,
		Appointmentdate: time.Now().UTC(),
		Duration:        duration.String(),
		Approval:        true,
	}
}

func Doctor_with_no_schedule_update(id int, patientid, appid int) models.Appointment {
	schedules, _ := services.ScheduleService.FindbyDoctor(id)
	for _, schedule := range schedules {
		if schedule.Active {
			services.ScheduleService.Update(models.Schedule{
				Scheduleid: schedule.Scheduleid,
				Active:     false,
			})
		}
	}
	duration, _ := time.ParseDuration("1h")
	return models.Appointment{
		Appointmentid:   appid,
		Doctorid:        id,
		Patientid:       patientid,
		Appointmentdate: time.Now().UTC(),
		Duration:        duration.String(),
		Approval:        true,
	}
}
func outboundappointment(a models.Appointment) models.Appointment {
	return models.Appointment{
		Appointmentid:   a.Appointmentid,
		Doctorid:        a.Doctorid,
		Patientid:       a.Patientid,
		Appointmentdate: a.Appointmentdate,
		Duration:        a.Duration,
		Approval:        true,
		Outbound:        true,
	}
}

func appointment_out_of_docs_schedule() models.Appointment {
	doctor := RandDoctor()
	patient := RandPatient()
	CreateSchedule(doctor.Physicianid)
	duration, _ := time.ParseDuration("1h")
	return models.Appointment{
		Doctorid:        doctor.Physicianid,
		Patientid:       patient.Patientid,
		Appointmentdate: time.Date(2023, 12, 12, 02, 30, 0, 0, time.UTC),
		Duration:        duration.String(),
		Approval:        true,
	}
}

func newappointment_same_doc(id int) models.Appointment {
	patient := RandPatient()
	duration, _ := time.ParseDuration("1h")
	return models.Appointment{
		Doctorid:        id,
		Patientid:       patient.Patientid,
		Appointmentdate: time.Date(2023, 12, 12, 12, 30, 0, 0, time.UTC),
		Duration:        duration.String(),
		Approval:        true,
	}
}

func newappointment_same_patient(id int) models.Appointment {
	doctor := RandDoctor()
	CreateSchedule(doctor.Physicianid)
	duration, _ := time.ParseDuration("1h")
	return models.Appointment{
		Doctorid:        doctor.Physicianid,
		Patientid:       id,
		Appointmentdate: time.Date(2023, 12, 12, 12, 30, 0, 0, time.UTC),
		Duration:        duration.String(),
		Approval:        true,
	}
}

func TestDoctorBookAppointmentService(t *testing.T) {
	data := newappointment()
	appointment1 := newappointment()
	testcases := []struct {
		description string
		data        models.Appointment
		test        func(*testing.T, models.Appointment, error)
	}{
		{
			description: "Book Appointment",
			data:        data,
			test: func(t *testing.T, a models.Appointment, err error) {
				require.NoError(t, err)
				require.Equal(t, a.Patientid, data.Patientid)
			},
		},
		{
			description: "Booking with clashing appointments",
			data: models.Appointment{
				Doctorid:        data.Doctorid,
				Patientid:       appointment1.Patientid,
				Appointmentdate: time.Now().UTC(),
				Duration:        "1h",
				Approval:        true,
			},
			test: func(t *testing.T, a models.Appointment, err error) {
				require.EqualError(t, err, ErrTimeSlotAllocated.Error())
				require.Empty(t, a)
			},
		},
		{
			// testing outbound patinet wiht already existing appointments
			description: "Outbound Appointment",
			data:        outboundappointment(data),
			test: func(t *testing.T, a models.Appointment, err error) {
				require.NoError(t, err)
				require.Equal(t, a.Patientid, data.Patientid)
			},
		},
		{
			description: "Doctor with no existing schedule",
			data:        Doctor_with_no_schedule_appointment(),
			test: func(t *testing.T, a models.Appointment, err error) {
				require.EqualError(t, err, ErrInvalidSchedule.Error())
				require.Empty(t, a)
			},
		},
		{
			description: "Not within doctors schedule",
			data:        appointment_out_of_docs_schedule(),
			test: func(t *testing.T, a models.Appointment, err error) {
				require.EqualError(t, err, ErrNotWithinSchedule.Error())
				require.Empty(t, a)
			},
		},
		{
			description: "Booking a second appointment",
			data:        newappointment_same_doc(data.Doctorid),
			test: func(t *testing.T, a models.Appointment, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, a)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.description, func(t *testing.T) {
			appointment, err := services.DoctorBookAppointment(tc.data)
			tc.test(t, appointment, err)
		})
	}
}

func TestPatientBookAppointmentService(t *testing.T) {
	appointment := newappointment()
	appointment1 := newappointment()
	testcases := []struct {
		description string
		data        models.Appointment
		test        func(*testing.T, models.Appointment, error)
	}{
		{
			description: "Book Appointment",
			data:        appointment,
			test: func(t *testing.T, a models.Appointment, err error) {
				require.NoError(t, err)
				require.Equal(t, a.Patientid, appointment.Patientid)
			},
		},
		{
			description: "Booking with clashing appointments",
			data: models.Appointment{
				Doctorid:        appointment1.Doctorid,
				Patientid:       appointment.Patientid,
				Appointmentdate: time.Now().UTC(),
				Duration:        "1h",
				Approval:        true,
			},
			test: func(t *testing.T, a models.Appointment, err error) {
				require.EqualError(t, err, ErrTimeSlotAllocated.Error())
				require.Empty(t, a)
			},
		},
		{
			description: "Doctor with no existing schedule",
			data:        Doctor_with_no_schedule_appointment(),
			test: func(t *testing.T, a models.Appointment, err error) {
				require.EqualError(t, err, ErrInvalidSchedule.Error())
				require.Empty(t, a)
			},
		},
		{
			description: "Not within doctors schedule",
			data:        appointment_out_of_docs_schedule(),
			test: func(t *testing.T, a models.Appointment, err error) {
				require.EqualError(t, err, ErrNotWithinSchedule.Error())
				require.Empty(t, a)
			},
		},
		{
			description: "Booking a second appointment",
			data:        newappointment_same_patient(appointment.Patientid),
			test: func(t *testing.T, a models.Appointment, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, a)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.description, func(t *testing.T) {
			appointment, err := services.PatientBookAppointment(tc.data)
			tc.test(t, appointment, err)
		})
	}
}

func TestDoctorUpdateAppointmentService(t *testing.T) {
	newappointment := newappointment()
	appointment, err := services.AppointmentService.Create(newappointment)
	require.NotNil(t, appointment)
	require.NoError(t, err)
	testcases := []struct {
		description string
		data        models.Appointment
		update      models.Appointment
		test        func(*testing.T, models.Appointment, error)
	}{
		{
			description: "Book Appointment",
			data:        appointment,
			update: models.Appointment{
				Appointmentid:   appointment.Appointmentid,
				Doctorid:        appointment.Doctorid,
				Patientid:       appointment.Patientid,
				Appointmentdate: time.Now().UTC(),
				Duration:        appointment.Duration,
				Approval:        true,
			},
			test: func(t *testing.T, a models.Appointment, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, a)
			},
		},
		{
			description: "Not within doctors schedule",
			data:        appointment,
			update: models.Appointment{
				Appointmentid:   appointment.Appointmentid,
				Doctorid:        appointment.Doctorid,
				Patientid:       appointment.Patientid,
				Appointmentdate: time.Date(2023, 12, 12, 2, 30, 0, 0, time.UTC),
				Duration:        appointment.Duration,
				Approval:        true,
			},
			test: func(t *testing.T, a models.Appointment, err error) {
				require.EqualError(t, err, ErrNotWithinSchedule.Error())
			},
		},
		{
			description: "Booking with clashing appointments",
			data:        appointment,
			update: models.Appointment{
				Appointmentid:   0,
				Doctorid:        appointment.Doctorid,
				Patientid:       appointment.Patientid,
				Appointmentdate: time.Now().UTC(),
				Duration:        appointment.Duration,
				Approval:        true,
			},
			test: func(t *testing.T, a models.Appointment, err error) {
				require.EqualError(t, err, ErrTimeSlotAllocated.Error())
				require.Empty(t, a)
			},
		},
		{
			// testing outbound patinet wiht already existing appointments
			description: "Outbound Appointment",
			data:        appointment,
			update:      outboundappointment(appointment),
			test: func(t *testing.T, a models.Appointment, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, a)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.description, func(t *testing.T) {
			updappointment, err := services.UpdateappointmentbyDoctor(tc.update)
			tc.test(t, updappointment, err)
		})
	}
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
func TestPatientUpdateAppointmentService(t *testing.T) {
	appointment1 := newappointment()
	appointment, err := services.AppointmentService.Create(appointment1)
	require.NotNil(t, appointment)
	require.NoError(t, err)
	testcases := []struct {
		description string
		update      models.Appointment
		test        func(*testing.T, models.Appointment, error)
	}{
		{
			description: "Book Outbound Appointment",
			update: models.Appointment{
				Appointmentid:   appointment.Appointmentid,
				Doctorid:        appointment.Doctorid,
				Patientid:       appointment.Patientid,
				Appointmentdate: time.Now().UTC(),
				Duration:        appointment.Duration,
				Approval:        true,
				Outbound:        true,
			},
			test: func(t *testing.T, a models.Appointment, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, a)
			},
		},
		{
			description: "Booking with clashing appointments",
			update: models.Appointment{
				Appointmentid:   -1,
				Doctorid:        appointment.Doctorid,
				Patientid:       appointment.Patientid,
				Appointmentdate: time.Now().UTC(),
				Duration:        appointment.Duration,
				Approval:        true,
			},
			test: func(t *testing.T, a models.Appointment, err error) {
				require.EqualError(t, err, ErrTimeSlotAllocated.Error())
				require.Empty(t, a)
			},
		},
		{
			description: "Approved appointment",
			update: models.Appointment{
				Appointmentid:   appointment.Appointmentid,
				Doctorid:        appointment.Doctorid,
				Patientid:       appointment.Patientid,
				Appointmentdate: time.Now().UTC(),
				Duration:        appointment.Duration,
				Approval:        true,
				Outbound:        false,
			},
			test: func(t *testing.T, a models.Appointment, err error) {
				require.EqualError(t, err, "can't update an approved appointment")
				require.Empty(t, a)
			},
		},
		{
			description: "Outbound appointment which is non-existing",
			update: models.Appointment{
				Appointmentid:   -1,
				Doctorid:        appointment.Doctorid,
				Patientid:       appointment.Patientid,
				Appointmentdate: time.Now().UTC(),
				Duration:        appointment.Duration,
				Approval:        true,
				Outbound:        true,
			},
			test: func(t *testing.T, a models.Appointment, err error) {
				require.EqualError(t, err, sql.ErrNoRows.Error())
				require.Empty(t, a)
			},
		},
		{
			description: "Book  Appointment",
			update: models.Appointment{
				Appointmentid:   appointment.Appointmentid,
				Doctorid:        appointment.Doctorid,
				Patientid:       appointment.Patientid,
				Appointmentdate: time.Now().UTC(),
				Duration:        appointment.Duration,
				Approval:        false,
				Outbound:        false,
			},
			test: func(t *testing.T, a models.Appointment, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, a)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.description, func(t *testing.T) {
			updappointment, err := services.UpdateappointmentbyPatient(tc.update)
			tc.test(t, updappointment, err)
		})
	}
}
func TestIsTimeWithinSchedule(t *testing.T) {
	tc := [][3]int64{
		{6, 12, 6},
		{6, 12, 7},
		{6, 12, 8},
		{6, 12, 11},
		{6, 12, 10},
		{4, 23, 7},
	}
	fc := [][3]int64{
		{6, 12, 1},
		{6, 12, 2},
		{6, 12, 12},
		{6, 12, 5},
		{6, 12, 13},
		{6, 12, 14},
	}
	for _, c := range tc {
		ok := isTimeWithinSchedule(c[0], c[1], c[2])
		require.Equal(t, ok, true)
	}
	for _, c := range fc {
		ok := isTimeWithinSchedule(c[0], c[1], c[2])
		require.Equal(t, ok, false)
	}
}
