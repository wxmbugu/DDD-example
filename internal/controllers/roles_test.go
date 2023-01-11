package controllers

import (
	"testing"

	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
	"github.com/stretchr/testify/require"
)

func CreateRoles() models.Roles {
	return models.Roles{
		Role: utils.RandString(6),
	}
}

func UpdateRole(id int) models.Roles {
	return models.Roles{
		Roleid: id,
		Role:   utils.RandString(5),
	}
}

func TestCreateRole(t *testing.T) {
	roles := CreateRoles()
	role, err := controllers.Roles.Create(roles)
	require.NoError(t, err)
	require.Equal(t, roles.Role, role.Role)

}

func TestFindRole(t *testing.T) {
	roles := CreateRoles()
	role, err := controllers.Roles.Create(roles)
	require.NoError(t, err)
	require.Equal(t, roles.Role, role.Role)
	r, err := controllers.Roles.Find(role.Roleid)
	require.NoError(t, err)
	require.Equal(t, r, role)
}

func TestListRoles(t *testing.T) {
	for i := 0; i < 5; i++ {
		roles := CreateRoles()
		role, err := controllers.Roles.Create(roles)
		require.NoError(t, err)
		require.Equal(t, roles.Role, role.Role)
	}

	schedules, err := controllers.Roles.FindAll()
	require.NoError(t, err)
	for _, v := range schedules {
		require.NotNil(t, v)
		require.NotEmpty(t, v)
	}

}

func TestDeleteRole(t *testing.T) {
	roles := CreateRoles()
	role, err := controllers.Roles.Create(roles)
	require.NoError(t, err)
	require.Equal(t, roles.Role, role.Role)
	err = controllers.Roles.Delete(role.Roleid)
	require.NoError(t, err)
	rolez, err := controllers.Roles.Find(role.Roleid)
	require.Error(t, err)
	require.Empty(t, rolez)
}

func TestUpdateRole(t *testing.T) {
	roles := CreateRoles()
	role, err := controllers.Roles.Create(roles)
	require.NoError(t, err)
	require.Equal(t, roles.Role, role.Role)
	role1 := UpdateRole(role.Roleid)
	role2, err := controllers.Roles.Update(role1)
	require.NoError(t, err)
	require.NotEqual(t, role.Role, role2.Role)
}
