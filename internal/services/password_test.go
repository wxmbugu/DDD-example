package services

import (
	"testing"

	"github.com/patienttracker/internal/utils"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHashpassword(t *testing.T) {
	password, err := HashPassword(utils.RandString(6))
	require.NoError(t, err)
	require.NotNil(t, password)
}

func CheckHashpasswors(t *testing.T) {
	password := utils.RandString(10)
	hashpassword, err := HashPassword(password)
	require.NoError(t, err)
	require.NotNil(t, password)
	err = CheckPassword(hashpassword, password)
	require.NoError(t, err)
	password2 := utils.RandString(10)
	err = CheckPassword(hashpassword, password2)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())
}
