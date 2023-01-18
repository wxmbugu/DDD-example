package controllers

import (
	"testing"

	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
	"github.com/stretchr/testify/require"
)

func UpdatePerm(id int, t *testing.T) models.Permissions {
	r := CreateRole(t)
	return models.Permissions{
		Permissionid: id,
		Permission:   utils.RandString(5),
		Roleid:       r.Roleid,
	}
}

func CreatePerm(t *testing.T) models.Permissions {
	r := CreateRole(t)
	return models.Permissions{
		Permission: utils.RandString(5),
		Roleid:     r.Roleid,
	}
}
func TestCreatePermissions(t *testing.T) {
	perm := CreatePerm(t)
	nperm, err := controllers.Permissions.Create(perm)
	require.NoError(t, err)
	require.Equal(t, perm.Permission, nperm.Permission)
}

func TestFindPermission(t *testing.T) {
	perm := CreatePerm(t)
	nperm, err := controllers.Permissions.Create(perm)
	require.NoError(t, err)
	require.Equal(t, perm.Permission, nperm.Permission)
	fperm, err := controllers.Permissions.Find(nperm.Permissionid)
	require.NoError(t, err)
	require.Equal(t, nperm, fperm)
}

func TestListPermissions(t *testing.T) {
	for i := 0; i < 5; i++ {
		perm := CreatePerm(t)
		nperm, err := controllers.Permissions.Create(perm)
		require.NoError(t, err)
		require.Equal(t, perm.Permission, nperm.Permission)
	}

	perms, err := controllers.Permissions.FindAll()
	require.NoError(t, err)
	for _, v := range perms {
		require.NotNil(t, v)
		require.NotEmpty(t, v)
	}

}

func TestListPermissionsbyRoles(t *testing.T) {
	r := CreateRole(t)
	for i := 0; i < 5; i++ {
		perm := models.Permissions{
			Permission: utils.RandString(5),
			Roleid:     r.Roleid,
		}
		nperm, err := controllers.Permissions.Create(perm)
		require.NoError(t, err)
		require.Equal(t, perm.Permission, nperm.Permission)
	}

	perms, err := controllers.Permissions.FindbyRoleId(r.Roleid)
	require.NoError(t, err)
	for _, v := range perms {
		require.NotNil(t, v)
		require.NotEmpty(t, v)
	}
}

func TestDeletePermissions(t *testing.T) {
	perm := CreatePerm(t)
	nperm, err := controllers.Permissions.Create(perm)
	require.NoError(t, err)
	require.Equal(t, perm.Permission, nperm.Permission)
	err = controllers.Permissions.Delete(nperm.Permissionid)
	require.NoError(t, err)
	fperm, err := controllers.Permissions.Find(nperm.Permissionid)
	require.Error(t, err)
	require.Empty(t, fperm)
}

func TestUpdatePermissions(t *testing.T) {
	perm := CreatePerm(t)
	nperm, err := controllers.Permissions.Create(perm)
	require.NoError(t, err)
	require.Equal(t, perm.Permission, nperm.Permission)
	perm1 := UpdatePerm(nperm.Permissionid, t)
	perm2, err := controllers.Permissions.Update(perm1)
	require.NoError(t, err)
	require.NotEqual(t, nperm.Permission, perm2.Permission)
}
