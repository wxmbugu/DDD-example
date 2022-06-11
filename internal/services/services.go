package services

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/patienttracker/internal/controllers"
	"github.com/patienttracker/internal/models"
	"github.com/patienttracker/internal/utils"
)

type Service struct {
	DoctorService      models.Physicianrepository
	AppointmentService models.AppointmentRepository
}

func NewService() Service {
	conn, err := sql.Open("postgres", "postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	controllers := controllers.New(conn)
	return Service{
		DoctorService: controllers.Doctors,
	}
}

func (s *Service) SomeService() (models.Physician, error) {
	doc, err := s.DoctorService.Create(
		models.Physician{
			Username:        utils.RandUsername(5),
			Full_name:       utils.Randfullname(10),
			Email:           utils.RandEmail(4),
			Hashed_password: utils.RandString(6),
			Contact:         utils.RandContact(10),
		},
	)
	return doc, err
}
