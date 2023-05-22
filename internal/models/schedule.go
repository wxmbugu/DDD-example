package models

// Schedule model
type (
	//Schedule struct hold the column type found in our Schedule table
	Schedule struct {
		Scheduleid int
		Doctorid   int
		Starttime  string
		Endtime    string
		Active     bool
	}

	//UpdateSchedule repository that holds the schedule model methods
	Schedulerepositroy interface {
		Create(schedule Schedule) (Schedule, error)
		Find(id int) (Schedule, error)
		FindAll(Filters) ([]Schedule, *Metadata, error)
		FindbyDoctor(id int) ([]Schedule, error)
		Delete(id int) error
		Update(schedule Schedule) (Schedule, error)
	}
)
