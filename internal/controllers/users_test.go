package controllers

import (
	"testing"

	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
	"github.com/stretchr/testify/require"
)

// insert a new value in to roles table for testing users and permissions
func CreateRole(t *testing.T) models.Roles {
	role := CreateRoles()
	r, err := controllers.Roles.Create(role)
	require.NoError(t, err)
	return r
}

func UpdateUser(id int, t *testing.T) models.Users {
	r := CreateRole(t)
	return models.Users{
		Id:       id,
		Email:    utils.RandEmail(5),
		Password: utils.RandString(6),
		Roleid:   r.Roleid,
	}
}

func CreateUser(t *testing.T) models.Users {
	r := CreateRole(t)
	return models.Users{
		Email:    utils.RandEmail(5),
		Password: utils.RandString(6),
		Roleid:   r.Roleid,
	}
}
func TestCreateUser(t *testing.T) {
	user := CreateUser(t)
	nuser, err := controllers.Users.Create(user)
	require.NoError(t, err)
	require.Equal(t, user.Email, nuser.Email)

}

func TestFindUser(t *testing.T) {
	user := CreateUser(t)
	nuser, err := controllers.Users.Create(user)
	require.NoError(t, err)
	require.Equal(t, user.Email, nuser.Email)
	fuser, err := controllers.Users.Find(nuser.Id)
	require.NoError(t, err)
	require.Equal(t, nuser, fuser)
}

func TestListUser(t *testing.T) {
	for i := 0; i < 5; i++ {
		user := CreateUser(t)
		nuser, err := controllers.Users.Create(user)
		require.NoError(t, err)
		require.Equal(t, user.Email, nuser.Email)
	}

	users, err := controllers.Users.FindAll()
	require.NoError(t, err)
	for _, v := range users {
		require.NotNil(t, v)
		require.NotEmpty(t, v)
	}

}

func TestListUserbyRoles(t *testing.T) {
	r := CreateRole(t)
	for i := 0; i < 5; i++ {
		user := models.Users{
			Email:    utils.RandEmail(5),
			Password: utils.RandString(6),
			Roleid:   r.Roleid,
		}
		nuser, err := controllers.Users.Create(user)
		require.NoError(t, err)
		require.Equal(t, user.Email, nuser.Email)
	}

	users, err := controllers.Users.FindbyRoleId(r.Roleid)
	require.NoError(t, err)
	for _, v := range users {
		require.NotNil(t, v)
		require.NotEmpty(t, v)
	}
}

func TestDeleteUser(t *testing.T) {
	user := CreateUser(t)
	nuser, err := controllers.Users.Create(user)
	require.NoError(t, err)
	require.Equal(t, user.Email, nuser.Email)
	err = controllers.Users.Delete(nuser.Id)
	require.NoError(t, err)
	fuser, err := controllers.Users.Find(nuser.Id)
	require.Error(t, err)
	require.Empty(t, fuser)
}

func TestUpdateUser(t *testing.T) {
	user := CreateUser(t)
	nuser, err := controllers.Users.Create(user)
	require.NoError(t, err)
	require.Equal(t, user.Email, nuser.Email)
	user1 := UpdateUser(nuser.Id, t)
	user2, err := controllers.Users.Update(user1)
	require.NoError(t, err)
	require.NotEqual(t, nuser.Email, user2.Email)
}
