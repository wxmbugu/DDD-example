package controllers

import (
	"testing"
	"time"

	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
	"github.com/stretchr/testify/require"
)

var testdoc models.Physicianrepository

func RandDoctor() models.Physician {
	username := utils.RandUsername(6)
	email := utils.RandEmail(5)
	fname := utils.Randfullname(4)
	//date := utils.Randate()
	return models.Physician{
		Username:        username,
		Full_name:       fname,
		Email:           email,
		Hashed_password: utils.RandString(8),
	}
}

func RandUpdDoctor() models.UpdatePhysician {
	username := utils.RandUsername(6)
	//contact := utils.RandContact(10)
	email := utils.RandEmail(5)
	fname := utils.Randfullname(4)
	//date := utils.Randate()
	return models.UpdatePhysician{
		Username:            username,
		Full_name:           fname,
		Email:               email,
		Hashed_password:     utils.RandString(8),
		Password_changed_at: time.Now(),
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
			user, err := testdoc.Create(scenario.input)
			require.NoError(t, err)
			require.Equal(t, doc.Username, user.Username)
		})

	}
}

func TestFindDoc(t *testing.T) {
	doc := RandDoctor()
	user, err := testdoc.Create(doc)
	require.NoError(t, err)
	newdoc, err := testdoc.Find(user.Physicianid)
	require.NoError(t, err)
	require.NotEmpty(t, newdoc)
	require.Equal(t, newdoc, user)
}

func TestListDocs(t *testing.T) {
	for i := 0; i < 5; i++ {
		doc := RandDoctor()
		_, err := testdoc.Create(doc)
		require.NoError(t, err)
	}
	docs, err := testdoc.FindAll()
	require.NoError(t, err)
	for _, v := range docs {
		require.NotNil(t, v)
		require.NotEmpty(t, v)
		require.Equal(t, v.Username, v.Username)
	}

}

func TestDeleteDoc(t *testing.T) {
	doc := RandDoctor()
	newdoc, err := testdoc.Create(doc)
	require.NoError(t, err)
	err = testdoc.Delete(newdoc.Physicianid)
	require.NoError(t, err)
	user2, err := testqueries.Find(newdoc.Physicianid)
	require.Error(t, err)
	require.Empty(t, user2)
}

func TestUpdateDoc(t *testing.T) {
	doc := RandDoctor()
	user, err := testdoc.Create(doc)
	require.NoError(t, err)
	docupd := RandUpdDoctor()
	update, err := testdoc.Update(docupd, user.Physicianid)
	require.NoError(t, err)
	require.Equal(t, docupd.Email, update.Email)
}
