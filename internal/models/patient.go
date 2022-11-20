package models

// Patient model struct
import "time"

type (
	Patient struct {
		Patientid          int
		Username           string
		Full_name          string
		Email              string
		Dob                time.Time
		Contact            string
		Bloodgroup         string
		Hashed_password    string
		Password_change_at time.Time
		Created_at         time.Time
		//verified           bool
	}
	ListPatients struct {
		Limit  int
		Offset int
	}
)

// PatientRepository represent the Patient repository contract
type PatientRepository interface {
	Create(patient Patient) (Patient, error)
	Find(id int) (Patient, error)
	FindbyEmail(email string) (Patient, error)
	FindAll(ListPatients) ([]Patient, error)
	Delete(id int) error
	Update(patient Patient) (Patient, error)
}
