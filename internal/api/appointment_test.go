package api

import (
	"bytes"
	///	"encoding/json"
	"fmt"
	"log"
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

func newappointment(date time.Time, approval bool, id int) models.Appointment {
	return models.Appointment{
		Appointmentid:   utils.Randid(1, 1000),
		Doctorid:        id,
		Patientid:       utils.Randid(1, 1000),
		Appointmentdate: date,
		Duration:        "1h",
		Approval:        approval,
	}
}

func createappointment(t *testing.T, date time.Time, approval bool, id int) models.Appointment {
	data, err := testserver.Services.AppointmentService.Create(newappointment(date, approval, id))
	require.NoError(t, err)
	return data
}

func createactiveappointment(t *testing.T) models.Appointment {
	schedule := createschedule(t, true)
	appointmentdate, err := time.Parse("2006-01-02 15:04", "2022-01-02 12:04")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(schedule.Doctorid)
	activeappoitnment := createappointment(t, appointmentdate, true, schedule.Doctorid)
	return activeappoitnment
}

func TestCreateAppointmentbyDoctor(t *testing.T) {
	schedule := createschedule(t, true)
	appointment := newappointment(time.Now(), false, schedule.Doctorid)

	appointmentdate, err := time.Parse("2006-01-02 15:04", "2022-01-02 12:04")
	if err != nil {
		log.Fatal(err)
	}
	activeappoitnment := createactiveappointment(t)
	newappointment := newappointment(appointmentdate, true, schedule.Doctorid)
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
				AppointmentReq{
					Doctorid:        newappointment.Doctorid,
					Patientid:       newappointment.Patientid,
					Appointmentdate: "2022-01-02 15:04",
					Duration:        appointment.Duration,
					Approval:        "false",
				},
			).Bytes(),
			id: newappointment.Doctorid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},

			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				fmt.Println("Body", recorder.Body)
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Invalid Field",
			id:   newappointment.Doctorid,
			body: encodetobytes(appointment.Appointmentid).Bytes(),
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "No active schedule",
			body: encodetobytes(
				AppointmentReq{
					Doctorid:        utils.Randid(1, 1000),
					Patientid:       appointment.Patientid,
					Appointmentdate: appointment.Appointmentdate.String(),
					Duration:        appointment.Duration,
					Approval:        "false",
				},
			).Bytes(),
			id: appointment.Doctorid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				fmt.Println(recorder.Body)
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Unauthorized",
			body: encodetobytes(
				AppointmentReq{
					Doctorid:        utils.Randid(1, 1000),
					Patientid:       appointment.Patientid,
					Appointmentdate: appointment.Appointmentdate.String(),
					Duration:        appointment.Duration,
					Approval:        "false",
				},
			).Bytes(),
			id: appointment.Doctorid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				fmt.Println(recorder.Body)
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},

		{
			name: "Booked Schedule",
			body: encodetobytes(
				AppointmentReq{
					Doctorid:        activeappoitnment.Doctorid,
					Patientid:       activeappoitnment.Patientid,
					Appointmentdate: "2022-01-02 12:04",
					Duration:        activeappoitnment.Duration,
					Approval:        "true",
				},
			).Bytes(),
			id: activeappoitnment.Doctorid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				fmt.Println(recorder.Body)
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/appointment/doctor/%d", tc.id)
			req, err := http.NewRequest(http.MethodPost, path, bytes.NewBuffer(tc.body))
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			tc.setauth(t, req, testserver.Auth)
			testserver.Router.HandleFunc("/v1/appointment/doctor/{id:[0-9]+}", testserver.createappointmentbydoctor)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}
func TestCreateAppointmentbyPatient(t *testing.T) {
	schedule := createschedule(t, true)
	appointment := newappointment(time.Now(), false, schedule.Doctorid)

	appointmentdate, err := time.Parse("2006-01-02 15:04", "2022-01-02 12:04")
	if err != nil {
		log.Fatal(err)
	}
	activeappoitnment := createappointment(t, appointmentdate, true, schedule.Doctorid)
	newappointment := newappointment(appointmentdate, true, schedule.Doctorid)
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
				AppointmentReq{
					Doctorid:        newappointment.Doctorid,
					Patientid:       newappointment.Patientid,
					Appointmentdate: "2022-01-02 15:04",
					Duration:        appointment.Duration,
					Approval:        "false",
				},
			).Bytes(),
			id: newappointment.Patientid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},

			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				fmt.Println("Body", recorder.Body)
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Invalid Field",
			id:   newappointment.Patientid,
			body: encodetobytes(appointment.Appointmentid).Bytes(),
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "No active schedule",
			body: encodetobytes(
				AppointmentReq{
					Doctorid:        utils.Randid(1, 1000),
					Patientid:       appointment.Patientid,
					Appointmentdate: appointment.Appointmentdate.String(),
					Duration:        appointment.Duration,
					Approval:        "false",
				},
			).Bytes(),
			id: appointment.Patientid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				fmt.Println(recorder.Body)
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Unauthorized",
			body: encodetobytes(
				AppointmentReq{
					Doctorid:        utils.Randid(1, 1000),
					Patientid:       appointment.Patientid,
					Appointmentdate: appointment.Appointmentdate.String(),
					Duration:        appointment.Duration,
					Approval:        "false",
				},
			).Bytes(),
			id: appointment.Patientid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				fmt.Println(recorder.Body)
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},

		{
			name: "Booked Schedule",
			body: encodetobytes(
				AppointmentReq{
					Doctorid:        activeappoitnment.Doctorid,
					Patientid:       activeappoitnment.Patientid,
					Appointmentdate: "2022-01-02 12:04",
					Duration:        activeappoitnment.Duration,
					Approval:        "true",
				},
			).Bytes(),
			id: activeappoitnment.Patientid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				fmt.Println(recorder.Body)
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/appointment/patient/%d", tc.id)
			req, err := http.NewRequest(http.MethodPost, path, bytes.NewBuffer(tc.body))
			require.NoError(t, err)
			tc.setauth(t, req, testserver.Auth)
			rr := httptest.NewRecorder()
			testserver.Router.HandleFunc("/v1/appointment/patient/{id:[0-9]+}", testserver.createappointmentbypatient)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}
func TestFindAppointment(t *testing.T) {
	activeappointment := createactiveappointment(t)
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		id       int
		setauth  func(t *testing.T, request *http.Request, token auth.Token)
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   activeappointment.Appointmentid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				require.Equal(t, encodetobytes(activeappointment), recorder.Body)
			},
		},
		{
			name: "Unauthorized",
			id:   activeappointment.Appointmentid,
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
				require.NotEqual(t, encodetobytes(activeappointment).Bytes(), recorder.Body.Bytes())
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/appointment/%d", tc.id)
			req, err := http.NewRequest(http.MethodGet, path, nil)
			require.NoError(t, err)
			tc.setauth(t, req, testserver.Auth)
			rr := httptest.NewRecorder()
			testserver.Router.HandleFunc("/v1/appointment/{id:[0-9]+}", testserver.findappointment)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestFindAllAppointments(t *testing.T) {

	appointment := createactiveappointment(t)

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
			id:     appointment.Appointmentid,
			Limit:  1,
			Offset: 5,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},

			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:   "Unauthorized",
			id:     appointment.Appointmentid,
			Limit:  1,
			Offset: 5,
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
			id:     appointment.Appointmentid,
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
			path := "/v1/appointments"
			req := httptest.NewRequest(http.MethodGet, path, nil)
			q := req.URL.Query()
			q.Add("page_id", strconv.Itoa(tc.Limit))
			q.Add("page_size", strconv.Itoa(tc.Limit))
			req.URL.RawQuery = q.Encode()
			rr := httptest.NewRecorder()
			tc.setauth(t, req, testserver.Auth)
			testserver.Router.Use(testserver.authmiddleware)
			testserver.Router.HandleFunc(path, testserver.findallappointments)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestFindAllAppointmentsbyDoctor(t *testing.T) {
	appointment := createactiveappointment(t)
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		id       int
		setauth  func(t *testing.T, request *http.Request, token auth.Token)
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   appointment.Doctorid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Unauthorized",
			id:   appointment.Doctorid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		//v1/patient/{id:[0-9]+}/appoinmtents
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/doctor/%d/appoinmtents", tc.id)
			req, err := http.NewRequest(http.MethodGet, path, nil)
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			tc.setauth(t, req, testserver.Auth)
			testserver.Router.Use(testserver.authmiddleware)
			testserver.Router.HandleFunc("/v1/doctor/{id:[0-9]+}/appoinmtents", testserver.findallappointmentsbydoctor)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}
func TestFindAllAppointmentsbyPatient(t *testing.T) {

	appointment := createactiveappointment(t)
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		id       int
		setauth  func(t *testing.T, request *http.Request, token auth.Token)
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   appointment.Doctorid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Unauthorized",
			id:   appointment.Doctorid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		}}

	for _, tc := range testcases {
		//v1/patient/{id:[0-9]+}/appoinmtents
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/patient/%d/appointments", tc.id)
			req, err := http.NewRequest(http.MethodGet, path, nil)
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			tc.setauth(t, req, testserver.Auth)
			testserver.Router.Use(testserver.authmiddleware)
			testserver.Router.HandleFunc("/v1/patient/{id:[0-9]+}/appointments", testserver.findallappointmentsbypatient)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestDeleteAppointment(t *testing.T) {

	appointment := createactiveappointment(t)
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		id       int
		setauth  func(t *testing.T, request *http.Request, token auth.Token)
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   appointment.Appointmentid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{name: "Unauthorized",
			id: appointment.Appointmentid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/appointment/%d", tc.id)
			req, err := http.NewRequest(http.MethodDelete, path, nil)
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			tc.setauth(t, req, testserver.Auth)
			testserver.Router.HandleFunc("/v1/appointment/{id:[0-9]+}", testserver.findappointment)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestUpdateAppointmentbyDoctor(t *testing.T) {
	//schedule := createschedule(t, true)
	appointment := newappointment(time.Now(), false, utils.Randid(1, 200))

	activeappoitnment := createactiveappointment(t)
	activeappoitnment2 := createactiveappointment(t)
	//newappointment := newappointment(appointmentdate, true, schedule.Doctorid)
	//var b bytes.Buffer
	testcases := []struct {
		name          string
		id            int
		appointmentid int
		body          []byte
		setauth       func(t *testing.T, request *http.Request, token auth.Token)
		response      func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: encodetobytes(
				AppointmentReq{
					Doctorid:        activeappoitnment.Doctorid,
					Patientid:       activeappoitnment.Patientid,
					Appointmentdate: "2022-01-02 09:04",
					Duration:        appointment.Duration,
					Approval:        "false",
				},
			).Bytes(),
			id:            activeappoitnment.Doctorid,
			appointmentid: activeappoitnment.Appointmentid,
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
				AppointmentReq{
					Doctorid:        activeappoitnment.Doctorid,
					Patientid:       activeappoitnment.Patientid,
					Appointmentdate: "2022-01-02 09:04",
					Duration:        appointment.Duration,
					Approval:        "false",
				},
			).Bytes(),
			id:            activeappoitnment.Doctorid,
			appointmentid: activeappoitnment.Appointmentid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:          "Invalid Field",
			id:            activeappoitnment.Doctorid,
			appointmentid: activeappoitnment.Appointmentid,
			body:          encodetobytes(appointment.Appointmentid).Bytes(),
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "No active schedule",
			body: encodetobytes(
				AppointmentReq{
					Doctorid:        utils.Randid(1, 1000),
					Patientid:       appointment.Patientid,
					Appointmentdate: appointment.Appointmentdate.String(),
					Duration:        appointment.Duration,
					Approval:        "false",
				},
			).Bytes(),
			id:            appointment.Doctorid,
			appointmentid: appointment.Appointmentid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				fmt.Println(recorder.Body)
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},

		{
			name: "Booked time slot",
			body: encodetobytes(
				AppointmentReq{
					Doctorid:        activeappoitnment.Doctorid,
					Patientid:       activeappoitnment.Patientid,
					Appointmentdate: "2022-01-02 09:04",
					Duration:        activeappoitnment.Duration,
					Approval:        "true",
				},
			).Bytes(),
			id:            activeappoitnment.Doctorid,
			appointmentid: activeappoitnment2.Appointmentid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				fmt.Println("error", recorder.Body)
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := "/v1/appointment/doctor"
			req, err := http.NewRequest(http.MethodPost, path, bytes.NewBuffer(tc.body))
			require.NoError(t, err)
			q := req.URL.Query()
			q.Add("id", strconv.Itoa(tc.appointmentid))
			q.Add("doctorid", strconv.Itoa(tc.id))
			req.URL.RawQuery = q.Encode()
			rr := httptest.NewRecorder()
			tc.setauth(t, req, testserver.Auth)
			testserver.Router.HandleFunc(path, testserver.UpdateDoctorAppointment)
			//testserver.Router.HandleFunc("/v1/appointment/{patientid:[0-9]+}/{id:[0-9]+}", testserver.updateappointmentbyPatient)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestUpdateAppointmentbyPatient(t *testing.T) {
	//schedule := createschedule(t, true)
	appointment := newappointment(time.Now(), false, utils.Randid(1, 200))

	activeappoitnment := createactiveappointment(t)
	activeappoitnment2 := createactiveappointment(t)
	//newappointment := newappointment(appointmentdate, true, schedule.Doctorid)
	//var b bytes.Buffer
	testcases := []struct {
		name          string
		id            int
		appointmentid int
		body          []byte
		setauth       func(t *testing.T, request *http.Request, token auth.Token)
		response      func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:          "OK",
			id:            activeappoitnment.Patientid,
			appointmentid: activeappoitnment.Appointmentid,
			body: encodetobytes(
				AppointmentReq{
					Doctorid:        activeappoitnment.Doctorid,
					Patientid:       activeappoitnment.Patientid,
					Appointmentdate: "2022-01-02 09:04",
					Duration:        appointment.Duration,
					Approval:        "false",
				},
			).Bytes(),
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},

			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		}, {
			name:          "Unauthorized",
			id:            activeappoitnment.Patientid,
			appointmentid: activeappoitnment.Appointmentid,
			body: encodetobytes(
				AppointmentReq{
					Doctorid:        activeappoitnment.Doctorid,
					Patientid:       activeappoitnment.Patientid,
					Appointmentdate: "2022-01-02 09:04",
					Duration:        appointment.Duration,
					Approval:        "false",
				},
			).Bytes(),
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:          "Invalid Field",
			id:            activeappoitnment.Doctorid,
			appointmentid: activeappoitnment.Appointmentid,
			body:          encodetobytes(appointment.Appointmentid).Bytes(),
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", "user", time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Already Booked",
			body: encodetobytes(
				AppointmentReq{
					Doctorid:        activeappoitnment.Doctorid,
					Patientid:       activeappoitnment.Patientid,
					Appointmentdate: "2022-01-02 09:04",
					Duration:        activeappoitnment.Duration,
					Approval:        "true",
				},
			).Bytes(),
			id:            activeappoitnment.Patientid,
			appointmentid: activeappoitnment2.Appointmentid,
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
			path := "/v1/appointment/patient"
			req, err := http.NewRequest(http.MethodPost, path, bytes.NewBuffer(tc.body))
			require.NoError(t, err)
			q := req.URL.Query()
			q.Add("id", strconv.Itoa(tc.appointmentid))
			q.Add("patientid", strconv.Itoa(tc.id))
			req.URL.RawQuery = q.Encode()
			rr := httptest.NewRecorder()
			tc.setauth(t, req, testserver.Auth)
			testserver.Router.HandleFunc(path, testserver.updateappointmentbyPatient)
			//testserver.Router.HandleFunc("/v1/appointment/{patientid:[0-9]+}/{id:[0-9]+}", testserver.updateappointmentbyPatient)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}
