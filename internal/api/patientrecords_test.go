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

	"github.com/patienttracker/internal/auth"
	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
	"github.com/stretchr/testify/require"
)

func newrecords() models.Patientrecords {
	return models.Patientrecords{
		Recordid:     utils.Randid(1, 1000),
		Patienid:     utils.Randid(1, 1000),
		Doctorid:     utils.Randid(1, 1000),
		Date:         time.Now(),
		Diagnosis:    utils.RandString(16),
		Disease:      utils.RandString(5),
		Prescription: utils.RandString(4),
		Weight:       fmt.Sprintf(utils.RandContact(2), "kg"),
	}
}

func createrecords(t *testing.T) models.Patientrecords {
	data, err := testserver.Services.PatientRecordService.Create(newrecords())
	require.NoError(t, err)
	return data
}

func TestCreateRecords(t *testing.T) {
	record := newrecords()
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		body     []byte
		setauth  func(t *testing.T, request *http.Request, token auth.Token)
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: encodetobytes(
				RecordReq{
					Patienid:     record.Patienid,
					Doctorid:     record.Doctorid,
					Date:         record.Date,
					Diagnosis:    record.Diagnosis,
					Disease:      record.Disease,
					Prescription: record.Prescription,
					Weight:       record.Weight,
				},
			).Bytes(),
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Unauthorized",
			body: encodetobytes(
				RecordReq{
					Patienid:     record.Patienid,
					Doctorid:     record.Doctorid,
					Date:         record.Date,
					Diagnosis:    record.Diagnosis,
					Disease:      record.Disease,
					Prescription: record.Prescription,
					Weight:       record.Weight,
				},
			).Bytes(),
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "Invalid Field",
			body: encodetobytes(record.Recordid).Bytes(),
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, "/v1/record", bytes.NewBuffer(tc.body))
			rr := httptest.NewRecorder()
			tc.setauth(t, req, testserver.Auth)
			testserver.Router.HandleFunc("/v1/record", testserver.createpatientrecord)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestFindRecords(t *testing.T) {
	record := createrecords(t)
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		id       int
		setauth  func(t *testing.T, request *http.Request, token auth.Token)
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   record.Recordid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},

			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				require.Equal(t, encodetobytes(record), recorder.Body)
			},
		},
		{
			name: "Unauthorized",
			id:   utils.Randid(1, 200),
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "Not Found",
			id:   utils.Randid(1, 200),
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", record.Disease, time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				require.NotEqual(t, encodetobytes(record).Bytes(), recorder.Body.Bytes())
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/record/%d", tc.id)
			req, _ := http.NewRequest(http.MethodGet, path, nil)
			tc.setauth(t, req, testserver.Auth)
			rr := httptest.NewRecorder()
			testserver.Router.HandleFunc("/v1/record/{id:[0-9]+}", testserver.findpatientrecord)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestFindAllRecords(t *testing.T) {
	var records models.Patientrecords
	for i := 0; i < 5; i++ {
		records = createrecords(t)
	}
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		id       int
		Limit    int
		Offset   int
		setauth  func(t *testing.T, request *http.Request, token auth.Token)
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			id:     records.Recordid,
			Limit:  1,
			Offset: 5000,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:   "NoAuthorization",
			id:     records.Recordid,
			Limit:  1,
			Offset: 5000,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "No query params",
			id:   utils.Randid(1, 200),
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "Invalid Page ID",
			id:     records.Recordid,
			Limit:  -1,
			Offset: 5,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/v1/record/", nil)
			q := req.URL.Query()
			q.Add("page_id", strconv.Itoa(tc.Limit))
			q.Add("page_size", strconv.Itoa(tc.Limit))
			req.URL.RawQuery = q.Encode()
			rr := httptest.NewRecorder()
			tc.setauth(t, req, testserver.Auth)
			testserver.Router.HandleFunc("/v1/record/", testserver.findallpatientrecords)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestFindAllRecordsbyDoctor(t *testing.T) {

	record := createrecords(t)
	var records []models.Patientrecords
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		id       int
		setauth  func(t *testing.T, request *http.Request, token auth.Token)
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   record.Doctorid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Unauthorized",

			id: utils.Randid(1, 200),
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "No Dept",

			id: utils.Randid(1, 200),
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				json.Unmarshal(recorder.Body.Bytes(), &records)
				require.Empty(t, records)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/doctor/%d/records", tc.id)
			req, _ := http.NewRequest(http.MethodGet, path, nil)
			rr := httptest.NewRecorder()
			tc.setauth(t, req, testserver.Auth)
			testserver.Router.HandleFunc("/v1/doctor/{id:[0-9]+}/records", testserver.findallrecordsbydoctor)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}
func TestFindAllRecordsbyPatient(t *testing.T) {

	record := createrecords(t)
	var records []models.Patientrecords
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		id       int
		setauth  func(t *testing.T, request *http.Request, token auth.Token)
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   record.Patienid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Unauthorized",
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
			},
			id: utils.Randid(1, 200),
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "No Dept",
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			id: utils.Randid(1, 200),
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				json.Unmarshal(recorder.Body.Bytes(), &records)
				require.Empty(t, records)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/patient/%d/records", tc.id)
			req, err := http.NewRequest(http.MethodGet, path, nil)
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			tc.setauth(t, req, testserver.Auth)
			testserver.Router.HandleFunc("/v1/patient/{id:[0-9]+}/records", testserver.updatepatientrecords)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestDeleteRecord(t *testing.T) {

	record := createrecords(t)
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		id       int
		setauth  func(t *testing.T, request *http.Request, token auth.Token)
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   record.Recordid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Unauthorized",
			id:   record.Recordid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/record/%d", tc.id)
			req, err := http.NewRequest(http.MethodDelete, path, nil)
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			tc.setauth(t, req, testserver.Auth)
			testserver.Router.HandleFunc("/v1/record/{id:[0-9]+}", testserver.deletepatientrecord)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestUpdateRecord(t *testing.T) {
	var somerecord models.Patientrecords
	record := createrecords(t)
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		id       int
		body     []byte
		setauth  func(t *testing.T, request *http.Request, token auth.Token)
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: encodetobytes(
				RecordReq{
					Patienid:     record.Patienid,
					Doctorid:     record.Doctorid,
					Date:         record.Date,
					Diagnosis:    record.Diagnosis,
					Disease:      utils.RandString(10),
					Prescription: record.Prescription,
					Weight:       record.Weight,
				},
			).Bytes(),
			id: record.Recordid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				json.Unmarshal(recorder.Body.Bytes(), &somerecord)
				require.Equal(t, record.Recordid, somerecord.Recordid)
				require.NotEqual(t, record.Disease, somerecord.Disease)
			},
		},
		{
			name: "Unauthorized",
			body: encodetobytes(
				RecordReq{
					Patienid:     record.Patienid,
					Doctorid:     record.Doctorid,
					Date:         record.Date,
					Diagnosis:    record.Diagnosis,
					Disease:      utils.RandString(10),
					Prescription: record.Prescription,
					Weight:       record.Weight,
				},
			).Bytes(),
			id: record.Recordid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "Invalid Field",
			body: encodetobytes(record.Date).Bytes(),
			id:   record.Recordid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/record/%d", tc.id)
			req, err := http.NewRequest(http.MethodPost, path, bytes.NewBuffer(tc.body))
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			tc.setauth(t, req, testserver.Auth)
			testserver.Router.HandleFunc("/v1/record/{id:[0-9]+}", testserver.updatepatientrecords)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}
