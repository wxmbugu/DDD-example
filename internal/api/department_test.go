package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateDepartment(t *testing.T) {
	//var dept models.Department
	//var b bytes.Buffer
	reqbody := new(bytes.Buffer)
	json.NewEncoder(reqbody).Encode(DepartmentReq{
		Departmentname: "smmm",
	})
	req := httptest.NewRequest("POST", "/v1/department", bytes.NewBuffer(reqbody.Bytes()))
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(testserver.createdepartment)
	handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	fmt.Println(rr.Body)
}
