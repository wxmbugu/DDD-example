package inmem

import (
	"errors"
	"sync"

	"github.com/patienttracker/internal/models"
)

type Department struct {
	mu   sync.RWMutex
	data map[int]models.Department
}

func (d *Department) Create(dept models.Department) (models.Department, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.data[dept.Departmentid] = dept
	return d.data[dept.Departmentid], nil
}
func (d *Department) Find(id int) (models.Department, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if val, ok := d.data[id]; ok {
		return val, nil
	}
	return models.Department{}, errors.New(" not such schedule")
}

// offset shouldn't be greater than limit
func (d *Department) FindAll(data models.ListDepartment) ([]models.Department, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	c := make([]models.Department, data.Offset, data.Limit)
	for _, val := range d.data {
		c = append(c, val)
	}
	return c, nil
}

func (d *Department) FindbyName(name string) (models.Department, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	//c := make([]models.Department, 0)
	for _, val := range d.data {
		if val.Departmentname == name {
			return val, nil
		}
	}
	return models.Department{}, errors.New("no such department")
}

func (d *Department) Delete(id int) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.data, id)
	return nil
}

func (d *Department) Update(dept models.Department) (models.Department, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.data[dept.Departmentid] = dept
	return d.data[dept.Departmentid], nil
}
