package controllers

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/patienttracker/internal/db"
	"github.com/patienttracker/internal/models"
)

type Physician struct {
	db *sql.DB
}

/*
  Create(patient Patient) (Patient, error)
	Find(id int) (Patient, error)
	FindAll() ([]Patient, error)
	Delete(id int) error
	Update(patient UpdatePatient) (Patient, error)
*/
func NewPhysicianRepositry() Physician {
	dbconn, err := db.New()
	if err != nil {
		log.Fatal(err)
	}
	return Physician{
		db: dbconn,
	}
}

func (p Physician) Create(patient models.Patient) (models.Patient, error) {
	sqlStatement := `
  INSERT INTO patient (username,hashed_password,full_name,email,dob,contact,bloodgroup) 
  s.dbconn
  RETURNING *
  `
	err := p.db.QueryRow(sqlStatement, patient.Username, patient.Hashed_password,
		patient.Full_name, patient.Email, patient.Dob, patient.Contact, patient.Bloodgroup).Scan(
		&patient.Patientid,
		&patient.Username,
		&patient.Hashed_password,
		&patient.Full_name,
		&patient.Email,
		&patient.Dob,
		&patient.Contact,
		&patient.Bloodgroup,
		&patient.Password_change_at,
		&patient.Created_at)
	if err != nil {
		log.Fatal(err)
	}
	return patient, nil

}

func (p Physician) Find(id int) (models.Patient, error) {
	sqlStatement := `
  SELECT * FROM patient
  WHERE patient.patientid = $1 LIMIT 1
  `
	var patient models.Patient
	err := p.db.QueryRowContext(context.Background(), sqlStatement, id).Scan(
		&patient.Patientid,
		&patient.Username,
		&patient.Hashed_password,
		&patient.Full_name,
		&patient.Email,
		&patient.Dob,
		&patient.Contact,
		&patient.Bloodgroup,
		&patient.Password_change_at,
		&patient.Created_at,
	)
	if err != nil {
		log.Fatal(err)
	}
	return patient, nil
}

type ListPhyysiciant struct {
	Limit  int
	Offset int
}

func (p Physician) FindAll() ([]models.Patient, error) {
	sqlStatement := `
 SELECT patientid, username,full_name,email,dob,contact,bloodgroup,created_at FROM patient
 ORDER BY patientid
 LIMIT $1
  `
	rows, err := p.db.QueryContext(context.Background(), sqlStatement, 10)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var items []models.Patient
	for rows.Next() {
		var i models.Patient
		if err := rows.Scan(
			&i.Patientid,
			&i.Username,
			&i.Full_name,
			&i.Email,
			&i.Dob,
			&i.Contact,
			&i.Bloodgroup,
			&i.Created_at,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (p Physician) Delete(id int) error {
	sqlStatement := `DELETE FROM patient
  WHERE patient.patientid = $1
  `
	_, err := p.db.Exec(sqlStatement, id)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (p Physician) Update(patient models.UpdatePatient, id int) (models.Patient, error) {
	sqlStatement := `UPDATE patient
SET username = $2, full_name = $3, email = $4,dob=$5,contact=$6,bloodgroup=$7,hashed_password=$8,password_change_at=$9
WHERE id = $1
RETURNING patientid,full_name,username,dob,contact,bloodgroup;
  `
	var user models.Patient
	err := p.db.QueryRow(sqlStatement, id, patient.Username, patient.Full_name, patient.Email, patient.Dob, patient.Contact, patient.Bloodgroup, patient.Hashed_password, time.Now()).Scan(
		&user.Patientid,
		&user.Username,
		&user.Hashed_password,
		&user.Full_name,
		&user.Email,
		&user.Dob,
		&user.Contact,
		&user.Bloodgroup,
	)
	if err != nil {
		log.Fatal(err)
	}
	return user, nil
}
