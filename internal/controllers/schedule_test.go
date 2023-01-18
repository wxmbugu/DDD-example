package controllers

import (
	"github.com/patienttracker/internal/models"
	"github.com/stretchr/testify/require"
	"testing"
)

func CreateSchedule(id int) models.Schedule {

	return models.Schedule{
		Doctorid:  id,
		Starttime: "09:00",
		Endtime:   "16:00",
		Active:    false,
	}
}

func UpdateSchedule(id int) models.Schedule {
	return models.Schedule{
		Scheduleid: id,
		Starttime:  "08:00",
		Endtime:    "17:00",
		Active:     true,
	}
}

func TestCreateSchedule(t *testing.T) {
	doc := RandDoctor()
	doctor, _ := controllers.Doctors.Create(doc)
	schedule := CreateSchedule(doctor.Physicianid)
	schedul, err := controllers.Schedule.Create(schedule)
	require.NoError(t, err)
	require.Equal(t, schedul.Doctorid, schedule.Doctorid)

}

func TestFindSchedule(t *testing.T) {
	doc := RandDoctor()
	doctor, _ := controllers.Doctors.Create(doc)
	schedule := CreateSchedule(doctor.Physicianid)
	schedul, err := controllers.Schedule.Create(schedule)
	require.NoError(t, err)
	work, err := controllers.Schedule.Find(schedul.Scheduleid)
	require.NoError(t, err)
	require.Equal(t, work, schedul)
}

func TestFindScheduleByDoctor(t *testing.T) {
	var sched models.Schedule
	doc := RandDoctor()
	doctor, _ := controllers.Doctors.Create(doc)
	schedule := CreateSchedule(doctor.Physicianid)
	for i := 0; i < 5; i++ {
		sched, _ = controllers.Schedule.Create(schedule)
		require.NotEmpty(t, sched)
	}
	schedules, err := controllers.Schedule.FindbyDoctor(sched.Doctorid)
	require.NoError(t, err)
	for _, v := range schedules {
		require.NotNil(t, v)
		require.NotEmpty(t, v)
		require.Equal(t, doctor.Physicianid, v.Doctorid)
	}
}

func TestListSchedule(t *testing.T) {
	for i := 0; i < 5; i++ {
		doc := RandDoctor()
		doctor, _ := controllers.Doctors.Create(doc)
		schedule := CreateSchedule(doctor.Physicianid)
		_, err := controllers.Schedule.Create(schedule)
		require.NoError(t, err)
	}
	args := models.ListSchedules{
		Limit:  5,
		Offset: 0,
	}
	schedules, err := controllers.Schedule.FindAll(args)
	require.NoError(t, err)
	for _, v := range schedules {
		require.NotNil(t, v)
		require.NotEmpty(t, v)
		require.Equal(t, 5, len(schedules))
	}

}

func TestDeleteSchedule(t *testing.T) {
	doc := RandDoctor()
	doctor, _ := controllers.Doctors.Create(doc)
	schedule := CreateSchedule(doctor.Physicianid)
	schedul, err := controllers.Schedule.Create(schedule)
	require.NoError(t, err)
	err = controllers.Schedule.Delete(schedul.Scheduleid)
	require.NoError(t, err)
	work, err := controllers.Schedule.Find(schedule.Scheduleid)
	require.Error(t, err)
	require.Empty(t, work)
}

func TestUpdateSchedule(t *testing.T) {
	doc := RandDoctor()
	doctor, _ := controllers.Doctors.Create(doc)
	schedule := CreateSchedule(doctor.Physicianid)
	schedul, err := controllers.Schedule.Create(schedule)
	require.NoError(t, err)
	schedule1 := UpdateSchedule(schedul.Scheduleid)
	schedule2, err := controllers.Schedule.Update(schedule1)
	require.NoError(t, err)
	require.NotEqual(t, schedule2.Starttime, schedule2.Endtime)
}
