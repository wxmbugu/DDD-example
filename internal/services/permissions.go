package services

import (
	"fmt"
	"strings"

	"github.com/patienttracker/internal/models"
)

type Permissions string

var deliminator = ":"

const (
	Viewer Permissions = "viewer"
	Editor Permissions = "editor"
	Admin  Permissions = "admin"
)

var permissionsMap = map[string]Permissions{
	"admin":  Admin,
	"editor": Editor,
	"viewer": Viewer,
}

func Str_to_Permission(str string) Permissions {
	var permission = strings.ToLower(str)
	permissionsmap := make(map[string]Permissions)
	permissionsmap = permissionsMap
	value := permissionsmap[permission]
	return value
}

func (a Permissions) toString() string {
	return fmt.Sprintf("%s", a)
}

func (a Permissions) isValid() error {
	switch a {
	case Admin, Editor, Viewer:
		return nil
	}
	return ErrInvalidPermissions
}

// custom Permission Definition inherits from Permissions Enum
// i.e records:create this means a specific role is mapped to a domain or a specific resource <domain:permissions>
// e.g (record:create) domain/resource records is only limited  to permissins create
func (a Permissions) Define(domainname string, perm Permissions) string {
	return fmt.Sprintf(domainname + deliminator + perm.toString())
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
	var admin string
	for _, v := range permissions {
		if v.Permission == "admin" {
			admin = "admin"
		}
	}
	if admin == "" {
		return permission, ErrForbidden
	}
	permission, err = s.RbacService.PermissionsService.Create(permission)
	return permission, nil
}

// check if permissions are correct
func checkRequiredpermissions(requiredpermssion Permissions, availablepermission string) error {
	if requiredpermssion.toString() == availablepermission {
		return nil
	}
	return ErrForbidden
}

// check if domain/table are correct
func AssertDomain(domainname string, requiredomain string) error {
	if domainname == requiredomain {
		return nil
	}
	return ErrForbidden
}

// LookupPermissions() handles permissions to get table or domain permission i.e <(table/domain):permission> and system permissions i.e <permission>
// and assert permissions
func LookupPermissions(permission string, required Permissions) error {
	values := strings.Split(permission, deliminator)
	if len(values) == 2 {
		if err := checkRequiredpermissions(required, values[1]); err != nil {
			return err
		}
	}
	if err := checkRequiredpermissions(required, permission); err != nil {
		return err
	}
	return nil
}
