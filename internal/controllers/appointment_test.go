package controllers

import (
	"testing"

	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
	"github.com/stretchr/testify/require"
)

func CreateAppointment() models.Appointment {
	time := utils.Randate()
	patient := RandPatient()
	patient1, _ := controllers.Patient.Create(patient)
	physcian := RandDoctor()
	doc, _ := controllers.Doctors.Create(physcian)
	appointment, _ := controllers.Appointment.Create(models.Appointment{
		Patientid:       patient1.Patientid,
		Doctorid:        doc.Physicianid,
		Appointmentdate: time,
		Duration:        "1h",
		Approval:        false,
	})
	return appointment
}

func TestCreateNewAppointment(t *testing.T) {
	time := utils.Randate()
	patient := RandPatient()
	patient1, _ := controllers.Patient.Create(patient)
	physcian := RandDoctor()
	doc, _ := controllers.Doctors.Create(physcian)
	appointment, err := controllers.Appointment.Create(models.Appointment{
		Patientid:       patient1.Patientid,
		Doctorid:        doc.Physicianid,
		Appointmentdate: time,
	})
	require.NoError(t, err)
	require.Equal(t, appointment.Patientid, patient1.Patientid)
}

func TestFindAppointment(t *testing.T) {
	appointment := CreateAppointment()
	schedule, err := controllers.Appointment.Find(appointment.Appointmentid)
	require.NoError(t, err)
	require.NotEmpty(t, appointment)
	require.Equal(t, appointment.Appointmentdate, schedule.Appointmentdate)
}

func TestListAppointments(t *testing.T) {
	for i := 0; i < 5; i++ {
		CreateAppointment()

	}
	appointment, err := controllers.Appointment.FindAll()
	require.NoError(t, err)
	for _, v := range appointment {
		require.NotNil(t, v)
		require.NotEmpty(t, v)
	}

}

func TestListAppointmentsByDoctor(t *testing.T) {
	appointment := CreateAppointment()
	appointments, err := controllers.Appointment.FindAllByDoctor(appointment.Doctorid)
	require.NoError(t, err)
	require.NotEmpty(t, appointments)
	for _, v := range appointments {
		require.NotNil(t, v)
		require.NotEmpty(t, v)
		require.Equal(t, appointment.Doctorid, v.Doctorid)
	}

}

func TestListAppointmentsByPatient(t *testing.T) {
	var appointment models.Appointment
	time := utils.Randate()
	patient := RandPatient()
	patient1, _ := controllers.Patient.Create(patient)
	physcian := RandDoctor()
	doc, _ := controllers.Doctors.Create(physcian)
	model := models.Appointment{
		Patientid:       patient1.Patientid,
		Doctorid:        doc.Physicianid,
		Appointmentdate: time,
	}
	//var appointment models.Appointment
	for i := 0; i < 5; i++ {
		appointment, _ = controllers.Appointment.Create(model)
	}
	appointments, err := controllers.Appointment.FindAllByPatient(appointment.Patientid)
	require.NoError(t, err)
	require.NotEmpty(t, appointments)
	for _, v := range appointments {
		require.NotNil(t, v)
		require.NotEmpty(t, v)
		require.Equal(t, appointment.Patientid, v.Patientid)
	}

}

func TestDeleteAppointments(t *testing.T) {
	appointment := CreateAppointment()
	err := controllers.Appointment.Delete(appointment.Appointmentid)
	require.NoError(t, err)
	schedule, err := controllers.Appointment.Find(appointment.Appointmentid)
	require.Error(t, err)
	require.Empty(t, schedule)
}

func TestUpdateAppointment(t *testing.T) {
	appointment := CreateAppointment()
	updt := models.AppointmentUpdate{
		Appointmentid:   appointment.Appointmentid,
		Appointmentdate: utils.Randate(),
		Duration:        "2h",
		Approval:        true,
	}
	updatedtime, err := controllers.Appointment.Update(updt)
	require.NoError(t, err)
	require.NotEqual(t, appointment.Appointmentdate, updatedtime)
}
