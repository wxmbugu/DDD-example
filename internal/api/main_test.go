package api

import (
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/patienttracker/internal/inmem"
	"github.com/patienttracker/internal/services"
)

func mockservices() services.Service {
	mockstore := inmem.NewMockStore()
	return services.Service{
		DoctorService:        mockstore.DoctorMemStore,
		AppointmentService:   mockstore.AppointmentMemStore,
		ScheduleService:      mockstore.ScheduleMemStore,
		PatientService:       mockstore.PatientMemStore,
		DepartmentService:    mockstore.DepartmentMemStore,
		PatientRecordService: mockstore.RecordMemStore,
	}
}

var testserver *Server

func TestMain(m *testing.M) {
	testserver = NewServer(mockservices(), mux.NewRouter())
	os.Exit(m.Run())
}
