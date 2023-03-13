package services

import (
	"github.com/patienttracker/internal/models"
	"strings"
)

func (service *Service) CreateRole(rolename string, roleid int) (models.Roles, error) {
	var role models.Roles
	var err error
	permissions, err := service.RbacService.PermissionsService.FindbyRoleId(roleid)
	if err != nil {
		return role, err
	}
	if len(permissions) == 0 {
		return role, ErrNotAuthorized
	}
	for _, permission := range permissions {
		if strings.ToLower(permission.Permission) == Admin.toString() {
			role, err = service.RbacService.RolesService.Create(models.Roles{
				Role: rolename,
			})
			if err != nil {
				return role, err
			}
			return role, nil
		} else {
			return role, ErrNotAuthorized
		}
	}
	return role, nil
}
