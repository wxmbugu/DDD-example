package models

type (
	Roles struct {
		Roleid int
		Role   string
	}
	Permissions struct {
		Permissionid int
		Permission   string
		Roleid       int
	}
	Users struct {
		Id       int
		Email    string
		Password string
		Roleid   int
	}
)

type RolesRepository interface {
	Create(Roles) (Roles, error)
	Find(int) (Roles, error)
	FindbyRole(string) (Roles, error)
	FindAll() ([]Roles, error)
	Delete(int) error
	Update(Roles) (Roles, error)
}

type PermissionsRepository interface {
	Create(Permissions) (Permissions, error)
	Find(int) (Permissions, error)
	FindAll() ([]Permissions, error)
	FindbyRoleId(int) ([]Permissions, error)
	Delete(int) error
	Update(Permissions) (Permissions, error)
}

type UsersRepository interface {
	Create(Users) (Users, error)
	Find(int) (Users, error)
	FindAll() ([]Users, error)
	FindbyEmail(string) (Users, error)
	FindbyRoleId(int) ([]Users, error)
	Delete(int) error
	Update(Users) (Users, error)
}
