package controllers

import (
	"testing"
	"time"

	//	"time"

	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
	"github.com/stretchr/testify/require"
)

/*
Create(schedule Schedule) (Schedule, error)
		Find(id int) (Schedule, error)
		FindAll() ([]Schedule, error)
		FindbyDoctor(id int) ([]Schedule, error)
		Delete(id int) error
		Update(schedule UpdateSchedule, id int) (Schedule, error)
*/

func CreateSchedule(id int) models.Schedule {
	//starttime := time.Now().String()
	//timec, _ := time.ParseDuration("-8h")
	//endtime := time.Now().Local().Add(timec).String()
	return models.Schedule{
		Doctorid:  id,
		Type:      "monthly",
		Starttime: time.Now(),
		Endtime:   utils.Randate(),
		Active:    false,
	}
}

func UpdateSchedule() models.UpdateSchedule {
	//stime, _ := time.Parse(starttime, starttime)
	//etime, _ := time.Parse(endtime, starttime)
	//h, _ := time.ParseDuration("8")
	return models.UpdateSchedule{
		Type:      "daily",
		Starttime: utils.Randate(),
		Endtime:   utils.Randate(),
		Active:    true,
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
	schedules, err := controllers.Schedule.FindAll()
	require.NoError(t, err)
	for _, v := range schedules {
		require.NotNil(t, v)
		require.NotEmpty(t, v)
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
	schedule1 := UpdateSchedule()
	schedule2, err := controllers.Schedule.Update(schedule1, schedul.Scheduleid)
	require.NoError(t, err)
	require.NotEqual(t, schedule2.Starttime, schedule2.Endtime)
}
