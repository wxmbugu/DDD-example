package mock

import (
	"errors"

	"github.com/patienttracker/internal/models"
)

type Department struct {
	data map[int]models.Department
}

func (d *Department) Create(dept models.Department) (models.Department, error) {
	d.data[dept.Departmentid] = dept
	return d.data[dept.Departmentid], nil
}
func (d *Department) Find(id int) (models.Department, error) {
	if val, ok := d.data[id]; ok {
		return val, nil
	}
	return models.Department{}, errors.New(" not such schedule")
}

// offset shouldn't be greater than limit
func (d *Department) FindAll(data models.ListSchedules) ([]models.Department, error) {
	c := make([]models.Department, data.Offset, data.Limit)
	for _, val := range d.data {
		c = append(c, val)
	}
	return c, nil
}

func (d *Department) FindbyName(name string) ([]models.Department, error) {
	c := make([]models.Department, 0)
	for _, val := range d.data {
		if val.Departmentname == name {
			c = append(c, val)
		}
	}
	return c, nil
}

func (d *Department) Delete(id int) error {
	delete(d.data, id)
	return nil
}

func (d *Department) Update(dept models.Department) (models.Department, error) {
	d.data[dept.Departmentid] = dept
	return d.data[dept.Departmentid], nil
}
