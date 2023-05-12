package api

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/patienttracker/internal/auth"
	"github.com/patienttracker/internal/inmem"
	"github.com/patienttracker/internal/services"
	"github.com/stretchr/testify/require"
	// "github.com/patienttracker/pkg/logger"
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

func encodetobytes(data any) *bytes.Buffer {
	reqbody := new(bytes.Buffer)
	json.NewEncoder(reqbody).Encode(data)
	return reqbody
}

func TestRegexContact(t *testing.T) {
	tc := []string{
		"0728519100",
		"254728519100",
		"+254728519100",
		"0128519100",
		"254128519100",
		"+254128519100",
	}
	for _, c := range tc {
		ok := checkinputregexformat(c, contactregex)
		require.Equal(t, true, ok)
	}
	fc := []string{
		"07285191",
		"2547q28519100",
		"+25472851911#900",
		"2*54128519100",
	}
	for _, c := range fc {
		ok := checkinputregexformat(c, contactregex)
		require.Equal(t, false, ok)
	}
}

func TestWeightRegex(t *testing.T) {
	weight := []string{
		"12lbs",
		"13kgs",
	}

	for _, c := range weight {
		ok := checkinputregexformat(c, weightregex)
		require.Equal(t, true, ok)
	}
	fweight := []string{
		"lbs12",
		"kgs13",
	}

	for _, c := range fweight {
		ok := checkinputregexformat(c, weightregex)
		require.Equal(t, false, ok)
	}
}

func TestMain(m *testing.M) {
	token, _ := auth.PasetoMaker("YELLOW SUBMARINE, BLACK WIZARDRY")
	testserver = NewServer(mockservices(), mux.NewRouter())
	testserver.Auth = token
	os.Exit(m.Run())
}
