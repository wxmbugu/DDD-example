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
	"github.com/gorilla/mux"
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
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Invalid Field",
			body: encodetobytes(record.Recordid).Bytes(),
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/record", bytes.NewBuffer(tc.body))
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(testserver.createpatientrecord)
			handler.ServeHTTP(rr, req)
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
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   record.Recordid,
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				require.Equal(t, encodetobytes(record), recorder.Body)
			},
		},
		{
			name: "Not Found",
			id:   utils.Randid(1, 200),
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				require.NotEqual(t, encodetobytes(record).Bytes(), recorder.Body.Bytes())
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/record/", nil)
			vars := map[string]string{
				"id": strconv.Itoa(tc.id),
			}
			req = mux.SetURLVars(req, vars)
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(testserver.findpatientrecord)
			handler.ServeHTTP(rr, req)
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
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			id:     records.Recordid,
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
			id:     records.Recordid,
			Limit:  -1,
			Offset: 5,
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/record/", nil)
			q := req.URL.Query()
			q.Add("page_id", strconv.Itoa(tc.Limit))
			q.Add("page_size", strconv.Itoa(tc.Limit))
			req.URL.RawQuery = q.Encode()
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(testserver.findallpatientrecords)
			handler.ServeHTTP(rr, req)
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
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   record.Doctorid,

			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "No Dept",

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
			req := httptest.NewRequest(http.MethodGet, "/v1/doctor/{id:[0-9]+}/records", nil)
			vars := map[string]string{
				"id": strconv.Itoa(tc.id),
			}
			req = mux.SetURLVars(req, vars)
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(testserver.findallrecordsbydoctor)
			handler.ServeHTTP(rr, req)
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
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   record.Patienid,

			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "No Dept",

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
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   record.Recordid,
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/record/%d", tc.id)
			req, err := http.NewRequest(http.MethodDelete, path, nil)
			require.NoError(t, err)
			rr := httptest.NewRecorder()
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
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				json.Unmarshal(recorder.Body.Bytes(), &somerecord)
				require.Equal(t, record.Recordid, somerecord.Recordid)
				require.NotEqual(t, record.Disease, somerecord.Disease)
			},
		},
		{
			name: "Invalid Field",
			body: encodetobytes(record.Date).Bytes(),
			id:   record.Recordid,
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
			testserver.Router.HandleFunc("/v1/record/{id:[0-9]+}", testserver.updatepatientrecords)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}
