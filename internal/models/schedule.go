package models

import "time"

//Schedule model
type (
	//Schedule struct hold the column type found in our Schedule table
	Schedule struct {
		Scheduleid int
		Doctorid   int
		Type       string
		Starttime  time.Time
		Endtime    time.Time
		Active     bool
	}
	//Update schedule struct
	UpdateSchedule struct {
		Type      string
		Starttime time.Time
		Endtime   time.Time
		Active    bool
	}
	//UpdateSchedule repository that holds the schedule model methods
	Schedulerepositroy interface {
		Create(schedule Schedule) (Schedule, error)
		Find(id int) (Schedule, error)
		FindAll() ([]Schedule, error)
		FindbyDoctor(id int) ([]Schedule, error)
		Delete(id int) error
		Update(schedule UpdateSchedule, id int) (Schedule, error)
	}
)
