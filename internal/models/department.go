package models

type (
	Department struct {
		Departmentid   int
		Departmentname string
	}

	Departmentrepository interface {
		Create(name string) (Department, error)
		Find(id int) (Department, error)
		FindbyName(name string) (Department, error)
		FindAll(limit int, offset int) ([]Department, error)
		Delete(id int) error
		Update(deptname string, id int) (Department, error)
	}
)
