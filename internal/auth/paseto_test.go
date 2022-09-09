package auth

import (
	"testing"
	"time"

	"github.com/patienttracker/internal/utils"
	"github.com/stretchr/testify/require"
	//"golang.org/x/tools/godoc/util"
)

func TestCreatePasetoToken(t *testing.T) {
	token, err := PasetoMaker(utils.RandString(32))
	require.NoError(t, err)
	duration := time.Minute
	username := utils.RandString(6)
	accesstoken, err := token.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, accesstoken)
	payload, err := token.VerifyToken(accesstoken)

	require.NoError(t, err)
	require.NotEmpty(t, payload)
	require.Equal(t, payload.Username, username)
}

func TestVerifyPasetoToken(t *testing.T) {
	token, err := PasetoMaker(utils.RandString(32))
	require.NoError(t, err)
	duration := -time.Minute
	username := utils.RandString(6)
	accesstoken, err := token.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, accesstoken)
	payload, err := token.VerifyToken(accesstoken)

	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Empty(t, payload)

}
