package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	//	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	//	"github.com/patienttracker/internal/models"

	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
	"github.com/stretchr/testify/require"
)

func newpatient() models.Patient {
	return models.Patient{
		Patientid:       utils.Randid(1, 1000),
		Username:        utils.RandUsername(6),
		Full_name:       utils.Randfullname(5),
		Email:           utils.RandEmail(6),
		Dob:             utils.Randate(),
		Contact:         utils.RandContact(10),
		Bloodgroup:      utils.RandString(2),
		Hashed_password: utils.RandString(10),
		Created_at:      time.Now(),
	}
}

func createpatient(t *testing.T) models.Patient {
	data, err := testserver.Services.PatientService.Create(newpatient())
	require.NoError(t, err)
	return data
}

func TestCreatepatient(t *testing.T) {
	patient := newpatient()
	fmt.Println(patient.Email)
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		body     []byte
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: encodetobytes(
				Patientreq{
					Username:        patient.Username,
					Full_name:       patient.Full_name,
					Email:           patient.Email,
					Dob:             patient.Dob.String(),
					Contact:         patient.Contact,
					Bloodgroup:      patient.Bloodgroup,
					Hashed_password: patient.Hashed_password,
				},
			).Bytes(),
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Invalid Field",
			body: encodetobytes(patient.Bloodgroup).Bytes(),
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Invalid Email Field",
			body: encodetobytes(
				Patientreq{
					Username:        patient.Username,
					Full_name:       patient.Full_name,
					Email:           utils.RandString(6),
					Dob:             patient.Dob.String(),
					Contact:         patient.Contact,
					Bloodgroup:      patient.Bloodgroup,
					Hashed_password: patient.Hashed_password,
				},
			).Bytes(),
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Invalid Password Field Lenght",
			body: encodetobytes(
				Patientreq{
					Username:        patient.Username,
					Full_name:       patient.Full_name,
					Email:           utils.RandString(6),
					Dob:             patient.Dob.String(),
					Contact:         patient.Contact,
					Bloodgroup:      patient.Bloodgroup,
					Hashed_password: utils.RandString(6),
				},
			).Bytes(),
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/patient", bytes.NewBuffer(tc.body))
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(testserver.createpatient)
			handler.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestFindPatient(t *testing.T) {
	patient := createpatient(t)
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		id       int
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   patient.Patientid,
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				require.Equal(t, encodetobytes(patient), recorder.Body)
			},
		},
		{
			name: "Not Found",
			id:   utils.Randid(1, 200),
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				require.NotEqual(t, encodetobytes(patient).Bytes(), recorder.Body.Bytes())
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/patient/%d", tc.id)
			req, err := http.NewRequest(http.MethodGet, path, nil)
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			testserver.Router.HandleFunc("/v1/patient/{id:[0-9]+}", testserver.findpatient)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestFindAllPatients(t *testing.T) {
	var patients models.Patient
	for i := 0; i < 5; i++ {
		patients = createpatient(t)
	}
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		id       int
		Limit    int
		Offset   int
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			id:     patients.Patientid,
			Limit:  1,
			Offset: 5000,
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "No query params",
			id:   utils.Randid(1, 200),
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "Invalid Page ID",
			id:     patients.Patientid,
			Limit:  -1,
			Offset: 5,
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := "/v1/patient"
			req, err := http.NewRequest(http.MethodGet, path, nil)
			require.NoError(t, err)
			q := req.URL.Query()
			q.Add("page_id", strconv.Itoa(tc.Limit))
			q.Add("page_size", strconv.Itoa(tc.Limit))
			req.URL.RawQuery = q.Encode()
			rr := httptest.NewRecorder()
			testserver.Router.HandleFunc("/v1/patient", testserver.findallpatients)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestDeletePatient(t *testing.T) {

	patient := createpatient(t)
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		id       int
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   patient.Patientid,
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/patient/%d", tc.id)
			req, err := http.NewRequest(http.MethodDelete, path, nil)
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			testserver.Router.HandleFunc("/v1/patient/{id:[0-9]+}", testserver.deletepatient)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestUpdatePatient(t *testing.T) {
	var somepatient models.Patient
	patient := newpatient()
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		body     []byte
		id       int
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: encodetobytes(
				Patientreq{
					Username:        patient.Username,
					Full_name:       patient.Full_name,
					Email:           "myemail@gmail.com",
					Dob:             patient.Dob.String(),
					Contact:         patient.Contact,
					Bloodgroup:      patient.Bloodgroup,
					Hashed_password: patient.Hashed_password,
				},
			).Bytes(),
			id: patient.Patientid,
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				json.Unmarshal(recorder.Body.Bytes(), &somepatient)
				require.NotEqual(t, patient.Email, somepatient.Email)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/patient/%d", tc.id)
			req, err := http.NewRequest(http.MethodPost, path, bytes.NewBuffer(tc.body))
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			testserver.Router.HandleFunc("/v1/patient/{id:[0-9]+}", testserver.updatepatient)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}
