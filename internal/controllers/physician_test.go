package controllers

import (
	"testing"
	"time"

	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
	"github.com/stretchr/testify/require"
)

func RandDoctor() models.Physician {
	username := utils.RandUsername(6)
	email := utils.RandEmail(5)
	fname := utils.Randfullname(4)
	deptname, _ := controllers.Department.Create(models.Department{Departmentname: utils.RandString(6)})
	//date := utils.Randate()
	return models.Physician{
		Username:        username,
		Full_name:       fname,
		Email:           email,
		Hashed_password: utils.RandString(8),
		Contact:         utils.RandContact(10),
		Departmentname:  deptname.Departmentname,
	}
}

func RandUpdDoctor(id int) models.Physician {
	username := utils.RandUsername(6)
	//contact := utils.RandContact(10)
	email := utils.RandEmail(5)
	fname := utils.Randfullname(4)
	//date := utils.Randate()
	deptname, _ := controllers.Department.Create(models.Department{Departmentname: utils.RandString(6)})
	return models.Physician{
		Physicianid:         id,
		Username:            username,
		Full_name:           fname,
		Email:               email,
		Hashed_password:     utils.RandString(8),
		Password_changed_at: time.Now(),
		Contact:             utils.RandContact(10),
		Departmentname:      deptname.Departmentname,
	}
}

func TestCreateDoc(t *testing.T) {
	doc := RandDoctor()
	type patientTest struct {
		description   string
		input         models.Physician
		expectedError string
	}
	for _, scenario := range []patientTest{
		{
			description:   "Passes",
			input:         doc,
			expectedError: "no errors!",
		},
	} {
		t.Run(scenario.description, func(t *testing.T) {
			user, err := controllers.Doctors.Create(scenario.input)
			require.NoError(t, err)
			require.Equal(t, doc.Contact, user.Contact)
		})

	}
}

func TestFindDoc(t *testing.T) {
	doc := RandDoctor()
	user, err := controllers.Doctors.Create(doc)
	require.NoError(t, err)
	newdoc, err := controllers.Doctors.Find(user.Physicianid)
	require.NoError(t, err)
	require.NotEmpty(t, newdoc)
	require.Equal(t, newdoc, user)
}

func TestFindDocbyEmail(t *testing.T) {
	doc := RandDoctor()
	user, err := controllers.Doctors.Create(doc)
	require.NoError(t, err)
	newdoc, err := controllers.Doctors.FindbyEmail(user.Email)
	require.NoError(t, err)
	require.NotEmpty(t, newdoc)
	require.Equal(t, newdoc, user)
}
func TestFindDocbyDept(t *testing.T) {
	var user models.Physician
	deptname, _ := controllers.Department.Create(models.Department{Departmentname: utils.RandString(6)})
	for i := 0; i < 5; i++ {
		username := utils.RandUsername(6)
		email := utils.RandEmail(5)
		fname := utils.Randfullname(4)
		doc := models.Physician{
			Username:        username,
			Full_name:       fname,
			Email:           email,
			Hashed_password: utils.RandString(8),
			Contact:         utils.RandContact(10),
			Departmentname:  deptname.Departmentname,
		}
		user, _ = controllers.Doctors.Create(doc)
	}
	args := models.ListDoctorsbyDeptarment{
		Department: user.Departmentname,
		Limit:      5,
		Offset:     0,
	}
	newdoc, err := controllers.Doctors.FindDoctorsbyDept(args)
	require.NoError(t, err)
	require.NotEmpty(t, newdoc)
	for _, v := range newdoc {
		require.NotNil(t, v)
		require.NotEmpty(t, v)
		require.Equal(t, 5, len(newdoc))
		require.Equal(t, deptname.Departmentname, v.Departmentname)
	}
}

func TestListDocs(t *testing.T) {
	for i := 0; i < 5; i++ {
		doc := RandDoctor()
		_, err := controllers.Doctors.Create(doc)
		require.NoError(t, err)
	}
	args := models.ListDoctors{
		Limit:  5,
		Offset: 0,
	}
	docs, err := controllers.Doctors.FindAll(args)
	require.NoError(t, err)
	for _, v := range docs {
		require.NotNil(t, v)
		require.NotEmpty(t, v)
		require.Equal(t, v.Username, v.Username)
		require.Equal(t, 5, len(docs))
	}

}

func TestDeleteDoc(t *testing.T) {
	doc := RandDoctor()
	newdoc, err := controllers.Doctors.Create(doc)
	require.NoError(t, err)
	err = controllers.Doctors.Delete(newdoc.Physicianid)
	require.NoError(t, err)
	user2, err := controllers.Doctors.Find(newdoc.Physicianid)
	require.Error(t, err)
	require.Empty(t, user2)
}

func TestUpdateDoc(t *testing.T) {
	doc := RandDoctor()
	user, err := controllers.Doctors.Create(doc)
	require.NoError(t, err)
	docupd := RandUpdDoctor(user.Physicianid)
	update, err := controllers.Doctors.Update(docupd)
	require.NoError(t, err)
	require.Equal(t, docupd.Email, update.Email)
}
