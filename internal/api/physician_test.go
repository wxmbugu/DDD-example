package api

import (
	"bytes"
	"encoding/json"
	"fmt"

	// "log"
	"strconv"
	"time"

	//	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	//	"github.com/patienttracker/internal/models"

	"github.com/patienttracker/internal/auth"
	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/services"
	"github.com/patienttracker/internal/utils"
	"github.com/stretchr/testify/require"
)

func newdoctor(name string) models.Physician {
	return models.Physician{
		Physicianid:     utils.Randid(1, 1000),
		Username:        utils.RandUsername(6),
		Full_name:       utils.Randfullname(5),
		Email:           utils.RandEmail(6),
		Contact:         utils.RandContact(10),
		Departmentname:  name,
		Hashed_password: utils.RandString(10),
		Created_at:      time.Now(),
	}
}

func createdoctor(t *testing.T) models.Physician {
	department := createdepartment(t)
	data, err := testserver.Services.DoctorService.Create(newdoctor(department.Departmentname))
	require.NoError(t, err)
	return data
}

func createlogindoctor(t *testing.T, doc models.Physician) models.Physician {
	doc.Hashed_password, _ = services.HashPassword(doc.Hashed_password)
	data, err := testserver.Services.DoctorService.Create(doc)
	require.NoError(t, err)
	return data
}
func TestCreateDoctor(t *testing.T) {
	department := createdepartment(t)
	doc := newdoctor(department.Departmentname)
	//var b bytes.Buffer
	testcases := []struct {
		name string
		body []byte
		// setauth  func(t *testing.T, request *http.Request, token auth.Token)
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: encodetobytes(
				Doctorreq{
					Username:        doc.Username,
					Full_name:       doc.Full_name,
					Email:           doc.Email,
					Departmentname:  doc.Departmentname,
					Contact:         doc.Contact,
					Hashed_password: doc.Hashed_password,
				},
			).Bytes(),
			// setauth: func(t *testing.T, request *http.Request, token auth.Token) {
			// 	create_auth(t, request, token, "bearer", doc.Username, time.Minute)
			// },
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Invalid Field",
			body: encodetobytes(doc.Contact).Bytes(),
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Invalid Email Field",
			body: encodetobytes(
				Doctorreq{
					Username:        doc.Username,
					Full_name:       doc.Full_name,
					Email:           utils.RandString(4),
					Departmentname:  doc.Departmentname,
					Contact:         doc.Contact,
					Hashed_password: doc.Hashed_password,
				},
			).Bytes(),
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Invalid Password Field Lenght",
			body: encodetobytes(
				Doctorreq{
					Username:        doc.Username,
					Full_name:       doc.Full_name,
					Email:           doc.Email,
					Departmentname:  doc.Departmentname,
					Contact:         doc.Contact,
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
			req := httptest.NewRequest(http.MethodPost, "/v1/doctor", bytes.NewBuffer(tc.body))
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(testserver.createdoctor)
			handler.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestDoctorLogin(t *testing.T) {
	department := createdepartment(t)
	d := newdoctor(department.Departmentname)
	doctor := createlogindoctor(t, d)
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		body     []byte
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: encodetobytes(
				PatientLoginreq{
					Email:    doctor.Email,
					Password: d.Hashed_password,
				},
			).Bytes(),
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Invalid Field",
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Invalid Email Field",
			body: encodetobytes(
				PatientLoginreq{
					Email:    utils.RandString(6),
					Password: d.Hashed_password,
				},
			).Bytes(),
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Wrong Password",
			body: encodetobytes(
				PatientLoginreq{
					Email:    doctor.Email,
					Password: utils.RandString(6),
				},
			).Bytes(),
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "No Password",
			body: encodetobytes(
				PatientLoginreq{
					Email: doctor.Email,
				},
			).Bytes(),
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/doctor/login", bytes.NewBuffer(tc.body))
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(testserver.DoctorLogin)
			handler.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}
func TestFindDoctor(t *testing.T) {
	doc := createdoctor(t)
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		id       int
		setauth  func(t *testing.T, request *http.Request, token auth.Token)
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   doc.Physicianid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", doc.Username, time.Minute)
			},

			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				require.Equal(t, encodetobytes(doc), recorder.Body)
			},
		},
		{
			name: "Unauthorized",
			id:   doc.Physicianid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Unauthorized", doc.Username, time.Minute)
			},

			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},

		{
			name: "Not Found",
			id:   utils.Randid(1, 200),
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", doc.Username, time.Minute)
			},

			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				require.NotEqual(t, encodetobytes(doc).Bytes(), recorder.Body.Bytes())
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/doctor/%d", tc.id)
			req, err := http.NewRequest(http.MethodGet, path, nil)
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			tc.setauth(t, req, testserver.Auth)
			testserver.Router.HandleFunc("/v1/doctor/{id:[0-9]+}", testserver.finddoctor)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestFindAllDoctor(t *testing.T) {
	var doc models.Physician
	for i := 0; i < 5; i++ {
		doc = createdoctor(t)
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
			id:     doc.Physicianid,
			Limit:  1,
			Offset: 5000,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", doc.Username, time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Unauthorized",
			id:   utils.Randid(1, 200),
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "", doc.Username, time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},

		{
			name: "No query params",
			id:   utils.Randid(1, 200),
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", doc.Username, time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "Invalid Page ID",
			id:     doc.Physicianid,
			Limit:  -1,
			Offset: 5,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", doc.Username, time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := "/v1/doctors/"
			req, err := http.NewRequest(http.MethodGet, path, nil)
			require.NoError(t, err)
			q := req.URL.Query()
			q.Add("page_id", strconv.Itoa(tc.Limit))
			q.Add("page_size", strconv.Itoa(tc.Limit))
			req.URL.RawQuery = q.Encode()
			rr := httptest.NewRecorder()
			tc.setauth(t, req, testserver.Auth)
			testserver.Router.HandleFunc("/v1/doctors/", testserver.findalldoctors)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestDeleteDoctor(t *testing.T) {

	doc := createdoctor(t)
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		id       int
		setauth  func(t *testing.T, request *http.Request, token auth.Token)
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   doc.Physicianid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", doc.Username, time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Unauthorized",
			id:   doc.Physicianid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "", doc.Username, time.Minute)
			},
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/doctor/%d", tc.id)
			req, err := http.NewRequest(http.MethodDelete, path, nil)
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			tc.setauth(t, req, testserver.Auth)
			testserver.Router.HandleFunc("/v1/doctor/{id:[0-9]+}", testserver.deletedoctor)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestUpdateDoctor(t *testing.T) {
	var somedoctor models.Physician
	doctor := createdoctor(t)
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		body     []byte
		id       int
		setauth  func(t *testing.T, request *http.Request, token auth.Token)
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: encodetobytes(
				Doctorreq{
					Username:        doctor.Username,
					Full_name:       doctor.Full_name,
					Email:           "doc@gmail.com",
					Departmentname:  doctor.Departmentname,
					Contact:         doctor.Contact,
					Hashed_password: utils.RandString(8),
				},
			).Bytes(),
			id: doctor.Physicianid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
				setup_auth(t, request, token, "Bearer", doctor.Username, time.Minute)
			},

			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				json.Unmarshal(recorder.Body.Bytes(), &somedoctor)
				require.NotEqual(t, doctor.Email, somedoctor.Email)
			},
		},
		{
			name: "Unauthorized",
			body: encodetobytes(
				Doctorreq{
					Username:        doctor.Username,
					Full_name:       doctor.Full_name,
					Email:           "doc@gmail.com",
					Departmentname:  doctor.Departmentname,
					Contact:         doctor.Contact,
					Hashed_password: utils.RandString(8),
				},
			).Bytes(),
			id: doctor.Physicianid,
			setauth: func(t *testing.T, request *http.Request, token auth.Token) {
			},

			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/v1/doctor/%d", tc.id)
			req, err := http.NewRequest(http.MethodPost, path, bytes.NewBuffer(tc.body))
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			tc.setauth(t, req, testserver.Auth)
			testserver.Router.HandleFunc("/v1/doctor/{id:[0-9]+}", testserver.updatedoctor)
			testserver.Router.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}
