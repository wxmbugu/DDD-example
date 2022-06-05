package controllers

import (
	"context"
	"log"
	"time"

	"github.com/patienttracker/internal/db"
	"github.com/patienttracker/internal/models"
)

type PatientRecords struct {
	db db.Database
}

/*
  Create(patient Patient) (Patient, error)
	Find(id int) (Patient, error)
	FindAll() ([]Patient, error)
	Delete(id int) error
	Update(patient UpdatePatient) (Patient, error)
*/
func NewPatientRecordsRepositry() PatientRecords {
	dbconn, err := db.New()
	if err != nil {
		log.Fatal(err)
	}
	return PatientRecords{
		db: dbconn,
	}
}

func (p PatientRecords) Create(patient models.Patient) (models.Patient, error) {
	sqlStatement := `
  INSERT INTO patient (username,hashed_password,full_name,email,dob,contact,bloodgroup) 
  s.dbconn
  RETURNING *
  `
	err := p.db.Conn.QueryRow(sqlStatement, patient.Username, patient.Hashed_password,
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

func (p PatientRecords) Find(id int) (models.Patient, error) {
	sqlStatement := `
  SELECT * FROM patient
  WHERE patient.patientid = $1 LIMIT 1
  `
	var patient models.Patient
	err := p.db.Conn.QueryRowContext(context.Background(), sqlStatement, id).Scan(
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

type ListPatientRecords struct {
	Limit  int
	Offset int
}

func (p PatientRecords) FindAll() ([]models.Patient, error) {
	sqlStatement := `
 SELECT patientid, username,full_name,email,dob,contact,bloodgroup,created_at FROM patient
 ORDER BY patientid
 LIMIT $1
  `
	rows, err := p.db.Conn.QueryContext(context.Background(), sqlStatement, 10)
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

func (p PatientRecords) Delete(id int) error {
	sqlStatement := `DELETE FROM patient
  WHERE patient.patientid = $1
  `
	_, err := p.db.Conn.Exec(sqlStatement, id)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (p PatientRecords) Update(patient models.UpdatePatient) (models.Patient, error) {
	sqlStatement := `UPDATE patient
SET username = $2, full_name = $3, email = $4,dob=$5,contact=$6,bloodgroup=$7,hashed_password=$8,Password_change_at=$9
WHERE id = $1
RETURNING patientid,full_name,username,dob,contact,bloodgroup;
  `
	var user models.Patient
	err := p.db.Conn.QueryRow(sqlStatement, patient.Id, patient.Username, patient.Full_name, patient.Email, patient.Dob, patient.Contact, patient.Bloodgroup, patient.Hashed_password, time.Now()).Scan(
		&user.Patientid,
		&user.Username,
		&user.Hashed_password,
		&user.Full_name,
		&user.Email,
		&user.Dob,
		&user.Contact,
		&user.Bloodgroup,
		&user.Password_change_at,
	)
	if err != nil {
		log.Fatal(err)
	}
	return user, nil
}
