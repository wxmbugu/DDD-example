package controllers

import (
	"testing"

	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
	"github.com/stretchr/testify/require"
)

var testappointment models.AppointmentRepository

func CreateAppointment() models.Appointment {
	time := utils.Randate()
	patient := RandPatient()
	patient1, _ := testqueries.Create(patient)
	physcian := RandDoctor()
	doc, _ := testdoc.Create(physcian)
	appointment, _ := testappointment.Create(models.Appointment{
		Patientid:       patient1.Patientid,
		Doctorid:        doc.Physicianid,
		Appointmentdate: time,
	})
	return appointment
}

func TestCreateNewAppointment(t *testing.T) {
	time := utils.Randate()
	patient := RandPatient()
	patient1, _ := testqueries.Create(patient)
	physcian := RandDoctor()
	doc, _ := testdoc.Create(physcian)
	appointment, err := testappointment.Create(models.Appointment{
		Patientid:       patient1.Patientid,
		Doctorid:        doc.Physicianid,
		Appointmentdate: time,
	})
	require.NoError(t, err)
	require.Equal(t, appointment.Patientid, doc.Physicianid)
}

func TestFindAppointment(t *testing.T) {
	appointment := CreateAppointment()
	schedule, err := testappointment.Find(appointment.Appointmentid)
	require.NoError(t, err)
	require.NotEmpty(t, appointment)
	require.Equal(t, appointment.Appointmentdate, schedule.Appointmentdate)
}

func TestListAppointments(t *testing.T) {
	for i := 0; i < 5; i++ {
		CreateAppointment()

	}
	appointment, err := testqueries.FindAll()
	require.NoError(t, err)
	for _, v := range appointment {
		require.NotNil(t, v)
		require.NotEmpty(t, v)

	}

}

func TestDeleteAppointments(t *testing.T) {
	appointment := CreateAppointment()
	err := testappointment.Delete(appointment.Appointmentid)
	require.NoError(t, err)
	schedule, err := testappointment.Find(appointment.Appointmentid)
	require.Error(t, err)
	require.Empty(t, schedule)
}

func TestUpdateAppointment(t *testing.T) {
	appointment := CreateAppointment()
	time := utils.Randate()
	updatedtime, err := testappointment.Update(time, appointment.Appointmentid)
	require.NoError(t, err)
	require.NotEqual(t, appointment.Appointmentdate, updatedtime)
}
