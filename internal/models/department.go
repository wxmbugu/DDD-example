package models

type (
	Department struct {
		Departmentid   int
		Departmentname string
	}
	ListDepartment struct {
		Limit  int
		Offset int
	}
	Departmentrepository interface {
		Create(Department) (Department, error)
		Find(id int) (Department, error)
		FindbyName(name string) (Department, error)
		FindAll(ListDepartment) ([]Department, error)
		Delete(id int) error
		Update(Department) (Department, error)
	}
)
