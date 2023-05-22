package inmem

import (
	"errors"
	"sync"

	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
)

type Doctor struct {
	mu   sync.RWMutex
	data map[int]models.Physician
}

func (d *Doctor) Create(doc models.Physician) (models.Physician, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	doc.Physicianid = utils.Randid(1, 10000)
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

func (d *Doctor) Count() (int, error) {
	count := len(d.data)
	return count, nil
}

// offset shouldn't be greater than limit
func (d *Doctor) FindAll(data models.Filters) ([]models.Physician, *models.Metadata, error) {
	count := 0
	var metadata models.Metadata
	d.mu.RLock()
	defer d.mu.RUnlock()
	c := make([]models.Physician, data.Offset(), data.Limit())
	for _, val := range d.data {
		count++
		c = append(c, val)
	}
	metadata = models.CalculateMetadata(count, data.Page, data.PageSize)
	return c, &metadata, nil
}
func (d *Doctor) Filter(full_name string, departmentname string, filters models.Filters) ([]*models.Physician, *models.Metadata, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	c := make([]*models.Physician, filters.Offset(), filters.Limit())
	for _, val := range d.data {
		c = append(c, &val)
	}
	return c, &models.Metadata{}, nil
}
func (d *Doctor) FindDoctorsbyDept(dept string, doc models.Filters) ([]models.Physician, *models.Metadata, error) {
	count := 0
	var metadata models.Metadata
	d.mu.RLock()
	defer d.mu.RUnlock()
	c := make([]models.Physician, doc.Offset(), doc.Limit())
	for _, val := range d.data {
		if val.Departmentname == dept {
			c = append(c, val)
		}
	}
	metadata = models.CalculateMetadata(count, doc.Page, doc.PageSize)
	return c, &metadata, nil
}
func (d *Doctor) Delete(id int) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.data, id)
	return nil
}

func (d *Doctor) FindbyEmail(email string) (models.Physician, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	for _, val := range d.data {
		if val.Email == email {
			return val, nil
		}
	}
	return models.Physician{}, errors.New("patient not found")
}

func (d *Doctor) Update(doc models.Physician) (models.Physician, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.data[doc.Physicianid] = doc
	return d.data[doc.Physicianid], nil
}
