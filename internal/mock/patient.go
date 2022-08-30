package mock

import (
	"errors"

	"github.com/patienttracker/internal/models"
)

type Patient struct {
	data map[int]models.Patient
}

func (p *Patient) Create(patient models.Patient) (models.Patient, error) {
	p.data[patient.Patientid] = patient
	return p.data[patient.Patientid], nil
}
func (p *Patient) Find(id int) (models.Patient, error) {
	if val, ok := p.data[id]; ok {
		return val, nil
	}
	return models.Patient{}, errors.New("patient not found")
}

// offset shouldn't be greater than limit
func (p *Patient) FindAll(data models.ListPatients) ([]models.Patient, error) {
	c := make([]models.Patient, data.Offset, data.Limit)
	for _, val := range p.data {
		c = append(c, val)
	}
	return c, nil
}

func (p *Patient) Delete(id int) error {
	delete(p.data, id)
	return nil
}

func (p *Patient) Update(patient models.Patient) (models.Patient, error) {
	p.data[patient.Patientid] = patient
	return p.data[patient.Patientid], nil
}
