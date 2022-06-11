package controllers

import (
	"testing"

	//	"time"

	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestCreateDeptName(t *testing.T) {
	dept, err := controllers.Department.Create(utils.RandString(6))
	require.NoError(t, err)
	require.NotEmpty(t, dept)

}

func TestFindDepartmentbyId(t *testing.T) {
	dept, err := controllers.Department.Create(utils.RandString(6))
	require.NoError(t, err)
	require.NotEmpty(t, dept)
	dept1, err := controllers.Department.Find(dept.Departmentid)
	require.NoError(t, err)
	require.NotEmpty(t, dept1)
	require.Equal(t, dept, dept1)
}

func TestFindDepartmentbyName(t *testing.T) {
	dept, err := controllers.Department.Create(utils.RandString(6))
	require.NoError(t, err)
	require.NotEmpty(t, dept)
	dept1, err := controllers.Department.FindbyName(dept.Departmentname)
	require.NoError(t, err)
	require.NotEmpty(t, dept1)
	require.Equal(t, dept, dept1)
}

func TestListDepartments(t *testing.T) {
	var dept models.Department
	for i := 0; i < 5; i++ {
		dept, _ = controllers.Department.Create(utils.RandString(6))
		require.NotEmpty(t, dept)
	}
	depts, err := controllers.Department.FindAll()
	require.NoError(t, err)
	require.NotEmpty(t, depts)
	require.Equal(t, len(depts), 10)

}

func TestDeleteDepartment(t *testing.T) {
	dept, err := controllers.Department.Create(utils.RandString(6))
	require.NoError(t, err)
	require.NotEmpty(t, dept)
	err = controllers.Department.Delete(dept.Departmentid)
	require.NoError(t, err)
	dept1, err := controllers.Department.Find(dept.Departmentid)
	require.Error(t, err)
	require.Empty(t, dept1)
}

func TestUpdateDepartment(t *testing.T) {
	dept, err := controllers.Department.Create(utils.RandString(6))
	require.NoError(t, err)
	require.NotEmpty(t, dept)
	dept1, err := controllers.Department.Update(utils.RandString(6), dept.Departmentid)
	require.NoError(t, err)
	require.NotEmpty(t, dept1)
	require.NotEqual(t, dept1.Departmentname, dept.Departmentname)
	require.Equal(t, dept1.Departmentid, dept.Departmentid)
}
