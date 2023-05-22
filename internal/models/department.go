package models

type (
	Department struct {
		Departmentid   int
		Departmentname string
	}

	Departmentrepository interface {
		Create(Department) (Department, error)
		Find(id int) (Department, error)
		FindbyName(name string) (Department, error)
		FindAll(Filters) ([]Department, *Metadata, error)
		Delete(id int) error
		Update(Department) (Department, error)
	}
)
