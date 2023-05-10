package controllers

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
	"github.com/stretchr/testify/require"
)

func RandNurse() models.Nurse {
	username := utils.RandUsername(6)
	fname := utils.Randfullname(12)
	email := utils.RandEmail(4)
	return models.Nurse{
		Username:        username,
		Full_name:       fname,
		Email:           email,
		Hashed_password: utils.RandString(7),
	}
}
func RandUpdNurse(email string, id int) models.Nurse {
	username := utils.RandUsername(6)
	fname := utils.Randfullname(12)
	return models.Nurse{
		Id:              id,
		Username:        username,
		Full_name:       fname,
		Email:           email,
		Hashed_password: utils.RandString(7),
	}
}

func TestCreateNurse(t *testing.T) {
	nurse := RandNurse()
	type NurseTest struct {
		description   string
		input         models.Nurse
		expectedError string
	}
	for _, scenario := range []NurseTest{
		{
			description:   "create acoount",
			input:         nurse,
			expectedError: "no error",
		},
	} {
		t.Run(scenario.description, func(t *testing.T) {
			user, err := controllers.Nurse.Create(scenario.input)
			require.NoError(t, err)
			require.Equal(t, nurse.Username, user.Username)
		})
	}
}

func TestFindNurse(t *testing.T) {
	nurse := RandNurse()
	type NurseTest struct {
		description string
		test        func(*testing.T, models.Nurse, error)
		data        models.Nurse
	}
	user, _ := controllers.Nurse.Create(nurse)
	for _, scenario := range []NurseTest{
		{
			description: "Account existent",
			data:        user,
			test: func(t *testing.T, n models.Nurse, err error) {
				require.NoError(t, err)
				require.Equal(t, n.Username, user.Username)
			},
		},
		{
			description: "Non-existent Account",
			data:        RandNurse(),
			test: func(t *testing.T, n models.Nurse, err error) {
				require.Empty(t, n)
				require.EqualError(t, err, sql.ErrNoRows.Error())
			},
		},
	} {
		t.Run(scenario.description, func(t *testing.T) {
			user, err := controllers.Nurse.Find(scenario.data.Id)
			scenario.test(t, user, err)
		})
	}
}

func TestFindNursebyEmail(t *testing.T) {
	nurse := RandNurse()
	type NurseTest struct {
		description string
		test        func(models.Nurse, error)
		data        models.Nurse
	}
	user, _ := controllers.Nurse.Create(nurse)
	for _, scenario := range []NurseTest{
		{
			description: "Account existent",
			data:        user,
			test: func(n models.Nurse, err error) {
				require.NoError(t, err)
				require.Equal(t, n.Username, user.Username)
			},
		},
		{
			description: "Non-existent Account",
			data:        RandNurse(),
			test: func(n models.Nurse, err error) {
				require.Empty(t, n)
				require.EqualError(t, err, sql.ErrNoRows.Error())
			},
		},
	} {
		t.Run(scenario.description, func(t *testing.T) {
			data, err := controllers.Nurse.FindbyEmail(scenario.data.Email)
			scenario.test(data, err)
		})
	}
}

func TestListNurses(t *testing.T) {
	for i := 0; i < 5; i++ {
		nurse := RandNurse()
		_, err := controllers.Nurse.Create(nurse)
		require.NoError(t, err)
	}
	args := models.ListNurses{
		Limit:  5,
		Offset: 0,
	}
	patients, err := controllers.Nurse.FindAll(args)
	require.NoError(t, err)
	for _, v := range patients {
		require.NotNil(t, v)
		require.NotEmpty(t, v)
		require.Equal(t, v.Username, v.Username)
	}
	require.Equal(t, 5, len(patients))
}
func TestCountNurses(t *testing.T) {
	count, err := controllers.Nurse.Count()
	fmt.Println(count)
	require.NoError(t, err)
	require.NotEmpty(t, count)
}
func TestDeleteNurse(t *testing.T) {
	nurse := RandNurse()
	type NurseTest struct {
		description string
		test        func(error)
		data        models.Nurse
	}
	user, _ := controllers.Nurse.Create(nurse)
	for _, scenario := range []NurseTest{
		{
			description: "Delete Account",
			data:        user,
			test: func(err error) {
				require.NoError(t, err)
			},
		},
		{
			description: "Delete Non-existence Account",
			data:        RandNurse(),
			test: func(err error) {
				require.NoError(t, err)
			},
		},
	} {
		t.Run(scenario.description, func(t *testing.T) {
			err := controllers.Nurse.Delete(scenario.data.Id)
			scenario.test(err)
		})
	}
}

func TestUpdateNurse(t *testing.T) {
	nurse := RandNurse()
	user, err := controllers.Nurse.Create(nurse)
	require.NoError(t, err)
	type NurseTest struct {
		description string
		test        func(models.Nurse, error)
		data        models.Nurse
	}
	for _, scenario := range []NurseTest{
		{
			description: "Update Existent account",
			data:        user,
			test: func(n models.Nurse, err error) {
				require.NoError(t, err)
				require.NotEqual(t, n.Username, nurse.Username)
				require.Equal(t, n.Email, user.Email)
			},
		},
		{
			description: "Update Non-Existent account",
			data:        RandNurse(),
			test: func(n models.Nurse, err error) {
				require.Empty(t, n)
				require.EqualError(t, err, sql.ErrNoRows.Error())
			},
		},
	} {
		t.Run(scenario.description, func(t *testing.T) {
			upd, err := controllers.Nurse.Update(RandUpdNurse(scenario.data.Email, scenario.data.Id))
			scenario.test(upd, err)
		})
	}
}
