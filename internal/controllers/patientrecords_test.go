package controllers

import (
	"testing"

	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
	"github.com/stretchr/testify/require"
)

func RandPatientRecord() models.Patientrecords {
	patient := RandPatient()
	doc := RandDoctor()
	pat, _ := controllers.Patient.Create(patient)
	physician, _ := controllers.Doctors.Create(doc)
	return models.Patientrecords{
		Patienid:     pat.Patientid,
		Doctorid:     physician.Physicianid,
		Date:         utils.Randate(),
		Diagnosis:    utils.RandString(20),
		Disease:      utils.RandString(10),
		Prescription: utils.RandString(6),
		Weight:       utils.RandContact(2) + "kgs",
	}
}

func RandUpdPatientrec() models.PatientrecordsUpd {
	return models.PatientrecordsUpd{
		Diagnosis:    utils.RandString(20),
		Disease:      utils.RandString(10),
		Prescription: utils.RandString(6),
		Weight:       utils.RandContact(2) + "kgs",
	}
}

func TestCreatePatientRecords(t *testing.T) {
	record := RandPatientRecord()
	precord, err := controllers.Records.Create(record)
	require.NoError(t, err)
	require.Equal(t, record.Diagnosis, precord.Diagnosis)
}

func TestFindPatientRecord(t *testing.T) {
	record := RandPatientRecord()
	precord, err := controllers.Records.Create(record)
	require.NoError(t, err)
	precord1, err := controllers.Records.Find(precord.Recordid)
	require.NoError(t, err)
	require.NotEmpty(t, precord1)
	require.Equal(t, precord, precord1)
}

func TestListPatientRecords(t *testing.T) {
	for i := 0; i < 5; i++ {
		record := RandPatientRecord()
		_, err := controllers.Records.Create(record)
		require.NoError(t, err)
	}
	records, err := controllers.Records.FindAll()
	require.NoError(t, err)
	for _, v := range records {
		require.NotNil(t, v)
		require.NotEmpty(t, v)
	}

}

func TestListPatientRecordsbyDoctor(t *testing.T) {
	record := RandPatientRecord()
	for i := 0; i < 5; i++ {

		_, err := controllers.Records.Create(record)
		require.NoError(t, err)
	}
	records, err := controllers.Records.FindAllByDoctor(record.Doctorid)
	require.NoError(t, err)
	for _, v := range records {
		require.NotNil(t, v)
		require.NotEmpty(t, v)
		require.Equal(t, record.Doctorid, v.Doctorid)
	}

}

func TestListPatientRecordsbyPatient(t *testing.T) {
	record := RandPatientRecord()
	for i := 0; i < 5; i++ {

		_, err := controllers.Records.Create(record)
		require.NoError(t, err)
	}
	records, err := controllers.Records.FindAllByPatient(record.Patienid)
	require.NoError(t, err)
	for _, v := range records {
		require.NotNil(t, v)
		require.NotEmpty(t, v)
		require.Equal(t, record.Patienid, v.Patienid)
	}

}

func TestDeletePatientRecord(t *testing.T) {
	record := RandPatientRecord()
	precord, err := controllers.Records.Create(record)
	require.NoError(t, err)
	err = controllers.Records.Delete(precord.Recordid)
	require.NoError(t, err)
	precord1, err := controllers.Records.Find(precord.Recordid)
	require.Error(t, err)
	require.Empty(t, precord1)
}

func TestUpdatePatientRecord(t *testing.T) {
	record := RandPatientRecord()
	precord, err := controllers.Records.Create(record)
	require.NoError(t, err)
	nrecord := RandUpdPatientrec()
	update, err := controllers.Records.Update(nrecord, precord.Recordid)
	require.NoError(t, err)
	require.NotEqual(t, precord.Diagnosis, update.Diagnosis)
}
