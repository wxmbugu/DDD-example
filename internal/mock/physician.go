package mock

import (
	"errors"
	"sync"

	"github.com/patienttracker/internal/models"
)

type Doctor struct {
	mu   sync.RWMutex
	data map[int]models.Physician
}

func (d *Doctor) Create(doc models.Physician) (models.Physician, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.data[doc.Physicianid] = doc
	return d.data[doc.Physicianid], nil
}
func (d *Doctor) Find(id int) (models.Physician, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if val, ok := d.data[id]; ok {
		return val, nil
	}
	return models.Physician{}, errors.New("doctor not found")
}

// offset shouldn't be greater than limit
func (d *Doctor) FindAll(data models.ListDoctors) ([]models.Physician, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	c := make([]models.Physician, data.Offset, data.Limit)
	for _, val := range d.data {
		c = append(c, val)
	}
	return c, nil
}

func (d *Doctor) Delete(id int) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.data, id)
	return nil
}

func (d *Doctor) Update(doc models.Physician) (models.Physician, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.data[doc.Physicianid] = doc
	return d.data[doc.Physicianid], nil
}
