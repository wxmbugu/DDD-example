package api

import (
	"bytes"

	//	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	//	"github.com/patienttracker/internal/models"
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

func TestCreateDepartment(t *testing.T) {
	department := newdepartment()
	//var b bytes.Buffer
	testcases := []struct {
		name     string
		body     []byte
		response func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{name: "OK",
			body: encodetobytes(
				DepartmentReq{
					Departmentname: department.Departmentname,
				},
			),
			response: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			}},
		{
			name: "Invalid Field",
			body: encodetobytes(department.Departmentid),
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
