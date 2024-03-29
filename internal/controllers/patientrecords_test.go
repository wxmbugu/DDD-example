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
	nurse, _ := controllers.Nurse.Create(RandNurse())
	return models.Patientrecords{
		Patienid:    pat.Patientid,
		Doctorid:    physician.Physicianid,
		Nurseid:     nurse.Id,
		Date:        utils.Randate(),
		Height:      utils.Randid(1, 100),
		Bp:          "100/200",
		HeartRate:   utils.Randid(1, 100),
		Temperature: utils.Randid(1, 37),
		Additional:  utils.RandString(100),
		Weight:      utils.RandContact(2) + "kgs",
	}
}

func RandUpdPatientrec(id int) models.Patientrecords {
	return models.Patientrecords{
		Recordid:    id,
		Height:      utils.Randid(1, 100),
		Bp:          "100/200",
		Temperature: utils.Randid(1, 37),
		Additional:  utils.RandString(100),
		Weight:      utils.RandContact(2) + "kgs",
	}
}

func TestCreatePatientRecords(t *testing.T) {
	record := RandPatientRecord()
	precord, err := controllers.Records.Create(record)
	require.NoError(t, err)
	require.Equal(t, record.Bp, precord.Bp)
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
	args := models.Filters{
		PageSize: 5,
		Page:     1,
	}
	records, _, err := controllers.Records.FindAll(args)
	require.NoError(t, err)
	for _, v := range records {
		require.NotNil(t, v)
		require.NotEmpty(t, v)
		require.Equal(t, 5, len(records))
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
	nrecord := RandUpdPatientrec(precord.Recordid)
	update, err := controllers.Records.Update(nrecord)
	require.NoError(t, err)
	require.NotEqual(t, precord.Weight, update.Weight)
}
