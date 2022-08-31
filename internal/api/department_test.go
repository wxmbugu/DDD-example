package api

import (
	"bytes"
	"encoding/json"
	"strconv"

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

func newdepartment() models.Department {
	return models.Department{
		Departmentid:   utils.Randid(1, 100),
		Departmentname: utils.RandString(6),
	}
}

func createdepartment(t *testing.T) models.Department {
	data, err := testserver.Services.DepartmentService.Create(newdepartment())
	require.NoError(t, err)
	return data
}

func TestCreateDepartment(t *testing.T) {
	department := newdepartment()
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		body     []byte
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: encodetobytes(
				DepartmentReq{
					Departmentname: department.Departmentname,
				},
			).Bytes(),
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Invalid Field",
			body: encodetobytes(department.Departmentid).Bytes(),
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/department", bytes.NewBuffer(tc.body))
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(testserver.createdepartment)
			handler.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestFindDepartment(t *testing.T) {
	department := createdepartment(t)
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		id       int
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   department.Departmentid,
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				require.Equal(t, encodetobytes(department), recorder.Body)
			},
		},
		{
			name: "Not Found",
			id:   utils.Randid(1, 200),
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				require.NotEqual(t, encodetobytes(department).Bytes(), recorder.Body.Bytes())
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/department/", nil)
			vars := map[string]string{
				"id": strconv.Itoa(tc.id),
			}
			req = mux.SetURLVars(req, vars)
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(testserver.finddepartment)
			handler.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestFindAllDepartments(t *testing.T) {
	var department models.Department
	for i := 0; i < 5; i++ {
		department = createdepartment(t)
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
			id:     department.Departmentid,
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
			id:     department.Departmentid,
			Limit:  -1,
			Offset: 5,
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
			handler := http.HandlerFunc(testserver.findalldepartment)
			handler.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestFindAllDcotorsbyDepartments(t *testing.T) {

	department := createdepartment(t)
	var departments []models.Department
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		deptname string
		Limit    int
		Offset   int
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "OK",
			deptname: department.Departmentname,
			Limit:    1,
			Offset:   5,
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:     "No Dept",
			Limit:    1,
			Offset:   5,
			deptname: utils.RandString(6),
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				json.Unmarshal(recorder.Body.Bytes(), &departments)
				require.Empty(t, departments)
			},
		},
		{
			name:     "Invalid Page ID",
			deptname: "ok",
			Limit:    -1,
			Offset:   5,
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/department/", nil)
			vars := map[string]string{
				"departmentname": tc.deptname,
			}
			req = mux.SetURLVars(req, vars)
			q := req.URL.Query()
			q.Add("page_id", strconv.Itoa(tc.Limit))
			q.Add("page_size", strconv.Itoa(tc.Limit))
			req.URL.RawQuery = q.Encode()
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(testserver.findalldoctorsbydepartment)
			handler.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestDeleteDepartment(t *testing.T) {
	department := createdepartment(t)
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		id       int
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   department.Departmentid,
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/v1/department/", nil)
			vars := map[string]string{
				"id": strconv.Itoa(tc.id),
			}
			req = mux.SetURLVars(req, vars)
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(testserver.deletedepartment)
			handler.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}

func TestUpdateDepartment(t *testing.T) {
	var somedepartment models.Department
	department := newdepartment()
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
				DepartmentReq{
					Departmentname: utils.RandString(6),
				},
			).Bytes(),
			id: department.Departmentid,
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				json.Unmarshal(recorder.Body.Bytes(), &somedepartment)
				require.Equal(t, department.Departmentid, somedepartment.Departmentid)
				require.NotEqual(t, department.Departmentname, somedepartment.Departmentname)
			},
		},
		{
			name: "Invalid Field",
			body: encodetobytes(department.Departmentid).Bytes(),
			id:   department.Departmentid,
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/department", bytes.NewBuffer(tc.body))
			vars := map[string]string{
				"id": strconv.Itoa(tc.id),
			}
			req = mux.SetURLVars(req, vars)
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(testserver.updatedepartment)
			handler.ServeHTTP(rr, req)
			tc.response(t, rr)
		})
	}
}
