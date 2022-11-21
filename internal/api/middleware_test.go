package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/patienttracker/internal/auth"
	"github.com/stretchr/testify/require"
)

func setup_auth(t *testing.T, request *http.Request, token auth.Token, authorizationType, username string, duration time.Duration) {
	accesstoken, err := token.CreateToken(username, duration)
	require.NoError(t, err)
	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, accesstoken)
	request.Header.Set(authHeaderKey, authorizationHeader)
}

func TestAuthmiddleware(t *testing.T) {
	testcases := []struct {
		name     string
		auth     func(t *testing.T, request *http.Request, token auth.Token)
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "Ok",
			auth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},

			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Expired token",
			auth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", -time.Minute)
			},

			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "Invalid header format",
			auth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "invalid", "user", time.Minute)
			},

			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "No auth header",
			auth: func(t *testing.T, request *http.Request, token auth.Token) {

			},

			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := "/v1/auth/"
			req, err := http.NewRequest(http.MethodGet, path, nil)
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			tc.auth(t, req, testserver.Auth)
			testserver.Router.Use(testserver.authmiddleware)
			testserver.Router.HandleFunc("/v1/auth/", func(w http.ResponseWriter, r *http.Request) {
				serializeResponse(w, http.StatusOK, "")
			})
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}
