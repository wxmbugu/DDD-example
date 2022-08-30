package mock

import (
	"errors"

	"github.com/patienttracker/internal/models"
)

type Schedule struct {
	data map[int]models.Schedule
}

func (s *Schedule) Create(schedule models.Schedule) (models.Schedule, error) {
	s.data[schedule.Scheduleid] = schedule
	return s.data[schedule.Scheduleid], nil
}
func (s *Schedule) Find(id int) (models.Schedule, error) {
	if val, ok := s.data[id]; ok {
		return val, nil
	}
	return models.Schedule{}, errors.New(" not such schedule")
}

// offset shouldn't be greater than limit
func (s *Schedule) FindAll(data models.ListSchedules) ([]models.Schedule, error) {
	c := make([]models.Schedule, data.Offset, data.Limit)
	for _, val := range s.data {
		c = append(c, val)
	}
	return c, nil
}

func (s *Schedule) FindbyDoctor(id int) ([]models.Schedule, error) {
	c := make([]models.Schedule, 0)
	for _, val := range s.data {
		if val.Doctorid == id {
			c = append(c, val)
		}
	}
	return c, nil
}

func (s *Schedule) Delete(id int) error {
	delete(s.data, id)
	return nil
}

func (s *Schedule) Update(schedule models.Schedule) (models.Schedule, error) {
	s.data[schedule.Scheduleid] = schedule
	return s.data[schedule.Scheduleid], nil
}
