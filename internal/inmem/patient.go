package inmem

import (
	"errors"
	"sync"

	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
)

type Patient struct {
	mu   sync.RWMutex
	data map[int]models.Patient
}

func (p *Patient) Create(patient models.Patient) (models.Patient, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	patient.Patientid = utils.Randid(1, 10000)
	p.data[patient.Patientid] = patient
	return p.data[patient.Patientid], nil
}
func (p *Patient) Find(id int) (models.Patient, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if val, ok := p.data[id]; ok {
		return val, nil
	}
	return models.Patient{}, errors.New("patient not found")
}
func (p *Patient) Count() (int, error) {
	count := len(p.data)
	return count, nil
}

func (p *Patient) FindbyEmail(email string) (models.Patient, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, val := range p.data {
		if val.Email == email {
			return val, nil
		}
	}
	return models.Patient{}, errors.New("patient not found")
}
func (d *Patient) Filter(full_name string, filters models.Filters) ([]*models.Patient, *models.Metadata, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	c := make([]*models.Patient, filters.Offset(), filters.Limit())
	for _, val := range d.data {
		c = append(c, &val)
	}
	return c, &models.Metadata{}, nil
}

// offset shouldn't be greater than limit
func (p *Patient) FindAll(data models.ListPatients) ([]models.Patient, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	c := make([]models.Patient, data.Offset, data.Limit)
	for _, val := range p.data {
		c = append(c, val)
	}
	return c, nil
}

func (p *Patient) Delete(id int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.data, id)
	return nil
}

func (p *Patient) Update(patient models.Patient) (models.Patient, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.data[patient.Patientid] = patient
	return p.data[patient.Patientid], nil
}
