package inmem

import (
	"errors"
	"sync"

	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
)

type Department struct {
	mu   sync.RWMutex
	data map[int]models.Department
}

func (d *Department) Create(dept models.Department) (models.Department, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	dept.Departmentid = utils.Randid(1, 10000)
	d.data[dept.Departmentid] = dept
	return d.data[dept.Departmentid], nil
}
func (d *Department) Find(id int) (models.Department, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if val, ok := d.data[id]; ok {
		return val, nil
	}
	return models.Department{}, errors.New(" no such department")
}
func (d *Department) Count() (int, error) {
	count := len(d.data)
	return count, nil
}

// offset shouldn't be greater than limit
func (d *Department) FindAll(data models.Filters) ([]models.Department, *models.Metadata, error) {
	count := 0
	var metadata models.Metadata
	d.mu.RLock()
	defer d.mu.RUnlock()
	c := make([]models.Department, data.Offset(), data.Limit())
	for _, val := range d.data {
		count++
		c = append(c, val)
	}
	metadata = models.CalculateMetadata(count, data.Page, data.PageSize)
	return c, &metadata, nil
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
