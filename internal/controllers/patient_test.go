package controllers

import (
	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func RandPatient() models.Patient {
	username := utils.RandUsername(6)
	contact := utils.RandContact(10)
	email := utils.RandEmail(5)
	fname := utils.Randfullname()
	date := utils.Randate()
	return models.Patient{
		Username:        username,
		Full_name:       fname,
		Email:           email,
		Dob:             date,
		Contact:         contact,
		Bloodgroup:      "A+",
		Hashed_password: utils.RandString(8),
	}
}

func RandUpdPatient(id int) models.Patient {
	username := utils.RandUsername(6)
	contact := utils.RandContact(10)
	email := utils.RandEmail(5)

	fname := utils.Randfullname()
	date := utils.Randate()
	return models.Patient{
		Patientid:          id,
		Username:           username,
		Full_name:          fname,
		Email:              email,
		Dob:                date,
		Contact:            contact,
		Bloodgroup:         "B+",
		Hashed_password:    utils.RandString(8),
		Password_change_at: time.Now(),
	}
}

func TestCreatePatient(t *testing.T) {
	patient := RandPatient()
	type patientTest struct {
		description   string
		input         models.Patient
		expectedError string
	}
	for _, scenario := range []patientTest{
		{
			description:   "create acoount",
			input:         patient,
			expectedError: "no error",
		},
	} {
		t.Run(scenario.description, func(t *testing.T) {
			user, err := controllers.Patient.Create(scenario.input)
			require.NoError(t, err)
			require.Equal(t, patient.Username, user.Username)
		})
	}
}

func TestFindPatient(t *testing.T) {
	patient := RandPatient()
	user, err := controllers.Patient.Create(patient)
	require.NoError(t, err)
	patient1, err := controllers.Patient.Find(user.Patientid)
	require.NoError(t, err)
	require.NotEmpty(t, patient)
	require.Equal(t, patient1.Email, user.Email)
}

func TestFindPatientbyEmail(t *testing.T) {
	patient := RandPatient()
	user, err := controllers.Patient.Create(patient)
	require.NoError(t, err)
	patient1, err := controllers.Patient.FindbyEmail(user.Email)
	require.NoError(t, err)
	require.NotEmpty(t, patient)
	require.Equal(t, patient1.Email, user.Email)
}

func TestListPatients(t *testing.T) {
	for i := 0; i < 5; i++ {
		patient := RandPatient()
		_, err := controllers.Patient.Create(patient)
		require.NoError(t, err)
	}
	args := models.Filters{
		PageSize: 5,
		Page:     1,
	}
	patients, _, err := controllers.Patient.FindAll(args)
	require.NoError(t, err)
	for _, v := range patients {
		require.NotNil(t, v)
		require.NotEmpty(t, v)
		require.Equal(t, v.Dob, v.Dob)
		require.Equal(t, v.Username, v.Username)
		require.Equal(t, 5, len(patients))
	}

}

func TestDeletePatient(t *testing.T) {
	patient := RandPatient()
	user, err := controllers.Patient.Create(patient)
	require.NoError(t, err)
	err = controllers.Patient.Delete(user.Patientid)
	require.NoError(t, err)
	user2, err := controllers.Patient.Find(user.Patientid)
	require.Error(t, err)
	require.Empty(t, user2)
}

func TestUpdatePatient(t *testing.T) {
	patient := RandPatient()
	user, err := controllers.Patient.Create(patient)
	require.NoError(t, err)
	patientupd := RandUpdPatient(user.Patientid)
	update, err := controllers.Patient.Update(patientupd)
	require.NoError(t, err)
	require.Equal(t, patientupd.Email, update.Email)
}
