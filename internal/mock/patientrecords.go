package mock

import (
	"errors"

	"github.com/patienttracker/internal/models"
)

type PatientRecords struct {
	data map[int]models.Patientrecords
}

func (p *PatientRecords) Create(records models.Patientrecords) (models.Patientrecords, error) {
	p.data[records.Recordid] = records
	return p.data[records.Recordid], nil
}
func (p *PatientRecords) Find(id int) (models.Patientrecords, error) {
	if val, ok := p.data[id]; ok {
		return val, nil
	}
	return models.Patientrecords{}, errors.New(" not such record")
}

// offset shouldn't be greater than limit
func (p *PatientRecords) FindAll(data models.ListAppointments) ([]models.Patientrecords, error) {
	c := make([]models.Patientrecords, data.Offset, data.Limit)
	for _, val := range p.data {
		c = append(c, val)
	}
	return c, nil
}

func (p *PatientRecords) FindAllbyDoctor(id int) ([]models.Patientrecords, error) {
	c := make([]models.Patientrecords, 0)
	for _, val := range p.data {
		if val.Doctorid == id {
			c = append(c, val)
		}
	}
	return c, nil
}

func (p *PatientRecords) FindAllByPatient(id int) ([]models.Patientrecords, error) {
	c := make([]models.Patientrecords, 0)
	for _, val := range p.data {
		if val.Patienid == id {
			c = append(c, val)
		}
	}
	return c, nil
}

func (p *PatientRecords) Delete(id int) error {
	delete(p.data, id)
	return nil
}

func (p *PatientRecords) Update(record models.Patientrecords) (models.Patientrecords, error) {
	p.data[record.Recordid] = record
	return p.data[record.Recordid], nil
}
