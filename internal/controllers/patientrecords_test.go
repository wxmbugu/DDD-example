package controllers

import (
	"testing"

	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
	"github.com/stretchr/testify/require"
)

var testrecord models.Patientrecordsrepository

func RandPatientRecord() models.Patientrecords {
	patient := RandPatient()
	doc := RandDoctor()
	pat, _ := testqueries.Create(patient)
	physician, _ := testdoc.Create(doc)
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
	precord, err := testrecord.Create(record)
	require.NoError(t, err)
	require.Equal(t, record.Diagnosis, precord.Diagnosis)
}

func TestFindPatientRecord(t *testing.T) {
	record := RandPatientRecord()
	precord, err := testrecord.Create(record)
	require.NoError(t, err)
	precord1, err := testrecord.Find(precord.Recordid)
	require.NoError(t, err)
	require.NotEmpty(t, precord1)
	require.Equal(t, precord, precord1)
}

func TestListPatientRecords(t *testing.T) {
	for i := 0; i < 5; i++ {
		record := RandPatientRecord()
		_, err := testrecord.Create(record)
		require.NoError(t, err)
	}
	records, err := testrecord.FindAll()
	require.NoError(t, err)
	for _, v := range records {
		require.NotNil(t, v)
		require.NotEmpty(t, v)
	}

}

func TestDeletePatientRecord(t *testing.T) {
	record := RandPatientRecord()
	precord, err := testrecord.Create(record)
	require.NoError(t, err)
	err = testrecord.Delete(precord.Recordid)
	require.NoError(t, err)
	precord1, err := testrecord.Find(precord.Recordid)
	require.Error(t, err)
	require.Empty(t, precord1)
}

func TestUpdatePatientRecord(t *testing.T) {
	record := RandPatientRecord()
	precord, err := testrecord.Create(record)
	require.NoError(t, err)
	nrecord := RandUpdPatientrec()
	update, err := testrecord.Update(nrecord, precord.Recordid)
	require.NoError(t, err)
	require.NotEqual(t, precord.Diagnosis, update.Diagnosis)
}
