package services

import (
	"fmt"

	"github.com/patienttracker/internal/models"
)

// import "sync"

type Permissions string

const (
	Create Permissions = "create"
	Read   Permissions = "read"
	Update Permissions = "update"
	Delete Permissions = "delete"
	CRUD   Permissions = "crud"
	Admin  Permissions = "admin"
)

type CreatePerm interface {
	Create(interface{}) interface{}
}

func (a Permissions) toString() string {
	return fmt.Sprintf("%s", a)
}

func (a Permissions) isValid() error {
	switch a {
	case Create, Read, Update, Delete, CRUD:
		return nil
	}
	return ErrInvalidPermissions
}

// patient:create
func (s *Service) CreatePermission(permission models.Permissions, userid int) (models.Permissions, error) {
	permissions, err := s.GetAllPermissionsofUser(userid)
	if err != nil {
		return models.Permissions{}, err
	}
	a := Permissions(permission.Permission)
	if err := a.isValid(); err != nil {
		return models.Permissions{}, err
	}
	for _, v := range permissions {
		if err := CheckRequiredpermissions("admin", v.Permission); err != nil {
			permission, err = s.RbacService.PermissionsService.Create(permission)
		}
	}
	return permission, err
}

func CheckRequiredpermissions(requiredpermssion string, availablepermission string) error {
	if requiredpermssion == availablepermission {
		return nil
	}
	return ErrForbidden
}
