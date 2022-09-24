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
	// "github.com/gorilla/mux"
	"github.com/patienttracker/internal/auth"
	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
	"github.com/stretchr/testify/require"
)

func newschedule(active bool) models.Schedule {
	return models.Schedule{
		Scheduleid: utils.Randid(1, 1000),
		Doctorid:   utils.Randid(1, 1000),
		Starttime:  "07:00",
		Endtime:    "23:00",
		Active:     active,
	}
}

func createschedule(t *testing.T, active bool) models.Schedule {
	data, err := testserver.Services.ScheduleService.Create(newschedule(active))
	require.NoError(t, err)
	return data
}

func TestCreateSchedule(t *testing.T) {
	schedule := createschedule(t, false)
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
				ScheduleReq{
					Doctorid:  schedule.Doctorid,
					Starttime: schedule.Starttime,
					Endtime:   schedule.Endtime,
					Active:    "true",
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
				ScheduleReq{
					Doctorid:  schedule.Doctorid,
					Starttime: schedule.Starttime,
					Endtime:   schedule.Endtime,
					Active:    "true",
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
			body: encodetobytes(schedule.Endtime).Bytes(),
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Schedule Exists",
			body: encodetobytes(
				ScheduleReq{
					Doctorid:  schedule.Doctorid,
					Starttime: schedule.Starttime,
					Endtime:   schedule.Endtime,
					Active:    "true",
				},
			).Bytes(),
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
			req, _ := http.NewRequest(http.MethodPost, "/v1/schedule", bytes.NewBuffer(tc.body))
			rr := httptest.NewRecorder()
			tc.setauth(t, req, testserver.Auth)
			testserver.Router.HandleFunc("/v1/schedule", testserver.createschedule)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestFindSchedule(t *testing.T) {
	schedule := createschedule(t, false)
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		id       int
		setauth  func(t *testing.T, request *http.Request, token auth.Token)
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   schedule.Scheduleid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				require.Equal(t, encodetobytes(schedule), recorder.Body)
			},
		},
		{
			name: "Unauthorized",
			id:   schedule.Scheduleid,
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
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				require.NotEqual(t, encodetobytes(schedule).Bytes(), recorder.Body.Bytes())
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/schedule/%d", tc.id)
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rr := httptest.NewRecorder()
			tc.setauth(t, req, testserver.Auth)
			testserver.Router.HandleFunc("/v1/schedule/{id:[0-9]+}", testserver.findschedule)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestFindAllSchedules(t *testing.T) {
	var schedule models.Schedule
	for i := 0; i < 5; i++ {
		schedule = createschedule(t, false)
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
			id:     schedule.Scheduleid,
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
			name:   "Unauthorized",
			id:     schedule.Scheduleid,
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
			id:     schedule.Scheduleid,
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
			req := httptest.NewRequest(http.MethodGet, "/v1/department/", nil)
			q := req.URL.Query()
			q.Add("page_id", strconv.Itoa(tc.Limit))
			q.Add("page_size", strconv.Itoa(tc.Limit))
			req.URL.RawQuery = q.Encode()
			rr := httptest.NewRecorder()
			tc.setauth(t, req, testserver.Auth)
			testserver.Router.HandleFunc("/v1/department/", testserver.findallschedules)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestFindAllSchedulesbyDoctor(t *testing.T) {
	var schedule models.Schedule
	for i := 0; i < 5; i++ {
		schedule = createschedule(t, false)
	}
	var schedules []models.Schedule
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		id       int
		setauth  func(t *testing.T, request *http.Request, token auth.Token)
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   schedule.Doctorid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Unauthorized",
			id:   schedule.Doctorid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "No Schedule",
			id:   utils.Randid(1, 100),
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				json.Unmarshal(recorder.Body.Bytes(), &schedules)
				require.Empty(t, schedules)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/doctor/%d/schedules", tc.id)
			req, err := http.NewRequest(http.MethodGet, path, nil)
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			tc.setauth(t, req, testserver.Auth)
			testserver.Router.HandleFunc("/v1/doctor/{id:[0-9]+}/schedules", testserver.findallschedulesbydoctor)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestDeleteSchedule(t *testing.T) {
	schedule := createschedule(t, false)
	testcases := []struct {
		name     string
		id       int
		setauth  func(t *testing.T, request *http.Request, token auth.Token)
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   schedule.Scheduleid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Unauthorized",
			id:   schedule.Scheduleid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/schedule/%d", tc.id)
			req, err := http.NewRequest(http.MethodDelete, path, nil)
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			tc.setauth(t, req, testserver.Auth)
			testserver.Router.HandleFunc("/v1/schedule/{id:[0-9]+}", testserver.deleteschedule)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestUpdateSchedule(t *testing.T) {
	var someschedule models.Schedule
	schedule := createschedule(t, true)
	inactiveschedule := createschedule(t, false)
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
				ScheduleReq{
					Doctorid:  schedule.Doctorid,
					Starttime: schedule.Starttime,
					Endtime:   "12:00",
					Active:    "true",
				},
			).Bytes(),
			id: schedule.Scheduleid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				json.Unmarshal(recorder.Body.Bytes(), &someschedule)
				require.Equal(t, schedule.Scheduleid, someschedule.Scheduleid)
				require.NotEqual(t, schedule.Endtime, someschedule.Endtime)
			},
		},
		{
			name: "Unauthorized",
			body: encodetobytes(
				ScheduleReq{
					Doctorid:  schedule.Doctorid,
					Starttime: schedule.Starttime,
					Endtime:   "12:00",
					Active:    "true",
				},
			).Bytes(),
			id: schedule.Scheduleid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},

		{
			name: "Invalid Field",
			body: encodetobytes(schedule.Scheduleid).Bytes(),
			id:   schedule.Scheduleid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Update only active schedule",
			body: encodetobytes(
				ScheduleReq{
					Doctorid:  schedule.Doctorid,
					Starttime: schedule.Starttime,
					Endtime:   schedule.Endtime,
					Active:    "true",
				},
			).Bytes(),
			id: inactiveschedule.Scheduleid,
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
			path := fmt.Sprintf("/v1/schedule/%d", tc.id)
			req, _ := http.NewRequest(http.MethodPost, path, bytes.NewBuffer(tc.body))
			rr := httptest.NewRecorder()
			tc.setauth(t, req, testserver.Auth)
			testserver.Router.HandleFunc("/v1/schedule", testserver.updateschedule)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}
