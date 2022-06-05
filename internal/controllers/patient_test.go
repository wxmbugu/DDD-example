package controllers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	patientrepo = NewPatientRepositry()
)

/*
func TestCreatePatient(t *testing.T) {
	patient1 := models.Patient{
		Username:        "Rayo",
		Full_name:       "Ryan mwaura",
		Hashed_password: "nekee",
		Email:           "rayo@gmail.com",
		Dob:             time.Now(),
		Contact:         "0710755945",
		Bloodgroup:      "A",
	}
	patientrepository := NewPatientRepositry()
	patient, err := patientrepository.Create(patient1)
	require.NoError(t, err)
	require.Equal(t, patient1.Full_name, patient.Full_name)
	require.Equal(t, patient1.Contact, patient.Contact)
	require.Equal(t, patient1.Email, patient.Email)
}
*/
func TestFindPatient(t *testing.T) {
	patient, err := patientrepo.Find(1)
	require.NoError(t, err)
	fmt.Println(patient)
	require.NotEmpty(t, patient)
}

func TestListPatients(t *testing.T) {
	patients, err := patientrepo.FindAll()
	require.NoError(t, err)
	fmt.Println(patients)
	for _, v := range patients {
		fmt.Println(v)
	}
	require.NotNil(t, patients)
	require.NotEmpty(t, patients)
}
