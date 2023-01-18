package inmem

import (
	"errors"
	"sync"

	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
)

type Appointment struct {
	mu   sync.RWMutex
	data map[int]models.Appointment
}

func (a *Appointment) Create(apntmnt models.Appointment) (models.Appointment, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	apntmnt.Appointmentid = utils.Randid(1, 10000)
	a.data[apntmnt.Appointmentid] = apntmnt
	return a.data[apntmnt.Appointmentid], nil
}
func (a *Appointment) Find(id int) (models.Appointment, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if val, ok := a.data[id]; ok {
		return val, nil
	}
	return models.Appointment{}, errors.New(" not such appointment")
}

// offset shouldn't be greater than limit
func (a *Appointment) FindAll(data models.ListAppointments) ([]models.Appointment, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	c := make([]models.Appointment, data.Offset, data.Limit)
	for _, val := range a.data {
		c = append(c, val)
	}
	return c, nil
}
func (a *Appointment) Count() (int, error) {
	count := len(a.data)
	return count, nil
}

func (a *Appointment) FindAllByDoctor(id int) ([]models.Appointment, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	c := make([]models.Appointment, 0)
	for _, val := range a.data {
		if val.Doctorid == id {
			c = append(c, val)
		}
	}
	return c, nil
}

func (a *Appointment) FindAllByPatient(id int) ([]models.Appointment, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	c := make([]models.Appointment, 0)
	for _, val := range a.data {
		if val.Patientid == id {
			c = append(c, val)
		}
	}
	return c, nil
}

func (a *Appointment) Delete(id int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.data, id)
	return nil
}

func (a *Appointment) Update(apntmnt models.Appointment) (models.Appointment, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.data[apntmnt.Appointmentid] = apntmnt
	return a.data[apntmnt.Appointmentid], nil
}
