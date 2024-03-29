package inmem

import (
	"errors"
	"sync"

	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
)

type PatientRecords struct {
	mu   sync.RWMutex
	data map[int]models.Patientrecords
}

func (p *PatientRecords) Create(records models.Patientrecords) (models.Patientrecords, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	records.Recordid = utils.Randid(1, 10000)
	p.data[records.Recordid] = records
	return p.data[records.Recordid], nil
}
func (p *PatientRecords) Find(id int) (models.Patientrecords, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if val, ok := p.data[id]; ok {
		return val, nil
	}
	return models.Patientrecords{}, errors.New(" not such record")
}

// offset shouldn't be greater than limit
func (p *PatientRecords) FindAll(data models.Filters) ([]models.Patientrecords, *models.Metadata, error) {
	count := 0
	var metadata models.Metadata
	p.mu.RLock()
	defer p.mu.RUnlock()
	c := make([]models.Patientrecords, data.Offset(), data.Limit())
	for _, val := range p.data {
		c = append(c, val)
	}
	metadata = models.CalculateMetadata(count, data.Page, data.PageSize)
	return c, &metadata, nil
}

func (p *PatientRecords) FindAllByDoctor(id int) ([]models.Patientrecords, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	c := make([]models.Patientrecords, 0)
	for _, val := range p.data {
		if val.Doctorid == id {
			c = append(c, val)
		}
	}
	return c, nil
}
func (p *PatientRecords) Count() (int, error) {
	count := len(p.data)
	return count, nil
}
func (p *PatientRecords) FindAllByPatient(id int) ([]models.Patientrecords, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	c := make([]models.Patientrecords, 0)
	for _, val := range p.data {
		if val.Patienid == id {
			c = append(c, val)
		}
	}
	return c, nil
}
func (p *PatientRecords) FindAllByNurse(id int) ([]models.Patientrecords, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	c := make([]models.Patientrecords, 0)
	for _, val := range p.data {
		if val.Nurseid == id {
			c = append(c, val)
		}
	}
	return c, nil
}
func (p *PatientRecords) Delete(id int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.data, id)
	return nil
}

func (p *PatientRecords) Update(record models.Patientrecords) (models.Patientrecords, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.data[record.Recordid] = record
	return p.data[record.Recordid], nil
}
