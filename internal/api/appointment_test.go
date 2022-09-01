package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
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
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		id       int
		body     []byte
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "Invalid Field",
			id:   newappointment.Doctorid,
			body: encodetobytes(appointment.Appointmentid).Bytes(),
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
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				fmt.Println(recorder.Body)
				require.Equal(t, http.StatusBadRequest, recorder.Code)
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
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				fmt.Println("Body", recorder.Body)
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Invalid Field",
			id:   newappointment.Patientid,
			body: encodetobytes(appointment.Appointmentid).Bytes(),
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
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				fmt.Println(recorder.Body)
				require.Equal(t, http.StatusBadRequest, recorder.Code)
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
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   activeappointment.Appointmentid,
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				require.Equal(t, encodetobytes(activeappointment), recorder.Body)
			},
		},
		{
			name: "Not Found",
			id:   utils.Randid(1, 200),
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
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			id:     appointment.Appointmentid,
			Limit:  1,
			Offset: 5,
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
			id:     appointment.Appointmentid,
			Limit:  -1,
			Offset: 5,
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := "/v1/appointments"
			req, err := http.NewRequest(http.MethodGet, path, nil)
			require.NoError(t, err)
			q := req.URL.Query()
			q.Add("page_id", strconv.Itoa(tc.Limit))
			q.Add("page_size", strconv.Itoa(tc.Limit))
			req.URL.RawQuery = q.Encode()
			rr := httptest.NewRecorder()
			testserver.Router.HandleFunc("/v1/appointments/", testserver.findallappointments)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestFindAllAppointmentsbyDoctor(t *testing.T) {
	appointment := createactiveappointment(t)
	var appointments []models.Appointment
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		id       int
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   appointment.Doctorid,

			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "No Dept",

			id: utils.Randid(1, 200),
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				json.Unmarshal(recorder.Body.Bytes(), &appointments)
				require.Empty(t, appointments)
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
			testserver.Router.HandleFunc("/v1/doctor/{id:[0-9]+}/appoinmtents", testserver.findallappointmentsbydoctor)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}
func TestFindAllAppointmentsbyPatient(t *testing.T) {

	appointment := createactiveappointment(t)
	var appointments []models.Appointment
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		id       int
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   appointment.Patientid,

			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "No Dept",

			id: utils.Randid(1, 200),
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				json.Unmarshal(recorder.Body.Bytes(), &appointments)
				require.Empty(t, appointments)
			},
		},
	}

	for _, tc := range testcases {
		//v1/patient/{id:[0-9]+}/appoinmtents
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/patient/%d/appoinmtents/", tc.id)
			req, err := http.NewRequest(http.MethodGet, path, nil)
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			testserver.Router.HandleFunc("/v1/patient/{id:[0-9]+}/appoinmtents/", testserver.findallappointmentsbypatient)
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
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   appointment.Appointmentid,
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/appointment/%d", tc.id)
			req, err := http.NewRequest(http.MethodDelete, path, nil)
			require.NoError(t, err)
			rr := httptest.NewRecorder()
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
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:          "Invalid Field",
			id:            activeappoitnment.Doctorid,
			appointmentid: activeappoitnment.Appointmentid,
			body:          encodetobytes(appointment.Appointmentid).Bytes(),
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
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				fmt.Println("error", recorder.Body)
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/appointment/%d/%d", tc.id, tc.appointmentid)
			req, err := http.NewRequest(http.MethodPatch, path, bytes.NewBuffer(tc.body))
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			testserver.Router.HandleFunc("/v1/appointment/{doctorid:[0-9]+}/{id:[0-9]+}", testserver.updateappointmentbyDoctor)
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
		response      func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		//{
		//	name: "OK",
		//	body: encodetobytes(
		//		AppointmentReq{
		//			Doctorid:        activeappoitnment.Doctorid,
		//			Patientid:       activeappoitnment.Patientid,
		//			Appointmentdate: "2022-01-02 09:04",
		//			Duration:        appointment.Duration,
		//			Approval:        "false",
		//		},
		//	).Bytes(),
		//	id:            activeappoitnment.Patientid,
		//	appointmentid: activeappoitnment.Appointmentid,
		//	response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		//		require.Equal(t, http.StatusOK, recorder.Code)
		//	},
		//},
		{
			name:          "Invalid Field",
			id:            activeappoitnment.Doctorid,
			appointmentid: activeappoitnment.Appointmentid,
			body:          encodetobytes(appointment.Appointmentid).Bytes(),
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
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/appointment/%d/%d", tc.id, tc.appointmentid)
			req, err := http.NewRequest(http.MethodPatch, path, bytes.NewBuffer(tc.body))
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			testserver.Router.HandleFunc("/v1/appointment/{patientid:[0-9]+}/{id:[0-9]+}", testserver.updateappointmentbyPatient)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}
