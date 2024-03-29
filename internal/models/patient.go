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
		Avatar             string
		About              string
		Created_at         time.Time
		Verified           bool
		Ischild            bool
	}
)

// PatientRepository represent the Patient repository contract
type PatientRepository interface {
	Create(patient Patient) (Patient, error)
	Find(id int) (Patient, error)
	FindbyEmail(email string) (Patient, error)
	Filter(string, Filters) ([]*Patient, *Metadata, error)
	FindAll(Filters) ([]Patient, *Metadata, error)
	Delete(id int) error
	Update(patient Patient) (Patient, error)
}
