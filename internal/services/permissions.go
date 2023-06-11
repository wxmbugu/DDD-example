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
	permission := strings.ToLower(str)
	return permissionsMap[permission]
}

func (a Permissions) toString() string {
	return string(a)
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
	return permission, err
}
