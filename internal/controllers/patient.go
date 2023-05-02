package controllers

import (
	"context"
	"database/sql"
	"log"

	"github.com/patienttracker/internal/models"
)

type Patient struct {
	db *sql.DB
}

/*
  Create(patient Patient) (Patient, error)
	Find(id int) (Patient, error)
	FindAll() ([]Patient, error)
	Delete(id int) error
	Update(patient UpdatePatient) (Patient, error)
*/

func (p Patient) Create(patient models.Patient) (models.Patient, error) {
	sqlStatement := `
  INSERT INTO patient (username,hashed_password,full_name,email,dob,contact,bloodgroup,about,verified,avatar,ischild) 
  VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
  RETURNING *;
  `
	err := p.db.QueryRow(sqlStatement, patient.Username, patient.Hashed_password,
		patient.Full_name, patient.Email, patient.Dob, patient.Contact, patient.Bloodgroup, patient.About, patient.Verified, patient.Avatar, patient.Ischild).Scan(
		&patient.Patientid,
		&patient.Username,
		&patient.Hashed_password,
		&patient.Full_name,
		&patient.Email,
		&patient.Dob,
		&patient.Contact,
		&patient.Bloodgroup,
		&patient.About,
		&patient.Verified,
		&patient.Avatar,
		&patient.Password_change_at,
		&patient.Created_at,
		&patient.Ischild)
	return patient, err

}

func (p Patient) Find(id int) (models.Patient, error) {
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
		&patient.About,
		&patient.Verified,
		&patient.Avatar,
		&patient.Password_change_at,
		&patient.Created_at,
		&patient.Ischild)
	return patient, err
}
func (p Patient) Count() (int, error) {

	counter := 0
	rows, err := p.db.Query("SELECT * FROM patient")
	if err != nil {
		return counter, err
	}
	defer rows.Close()

	for rows.Next() {
		counter++
	}
	return counter, nil
}
func (p Patient) FindbyEmail(email string) (models.Patient, error) {
	sqlStatement := `
  SELECT * FROM patient
  WHERE patient.email = $1 LIMIT 1
  `
	var patient models.Patient
	err := p.db.QueryRowContext(context.Background(), sqlStatement, email).Scan(
		&patient.Patientid,
		&patient.Username,
		&patient.Hashed_password,
		&patient.Full_name,
		&patient.Email,
		&patient.Dob,
		&patient.Contact,
		&patient.Bloodgroup,
		&patient.About,
		&patient.Verified,
		&patient.Avatar,
		&patient.Password_change_at,
		&patient.Created_at,
		&patient.Ischild)
	return patient, err
}

func (p Patient) FindAll(args models.ListPatients) ([]models.Patient, error) {
	sqlStatement := `
 SELECT patientid, username,full_name,email,dob,contact,bloodgroup,created_at FROM patient
 ORDER BY patientid
 LIMIT $1
 OFFSET $2
  `
	rows, err := p.db.QueryContext(context.Background(), sqlStatement, args.Limit, args.Offset)
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

func (p Patient) Delete(id int) error {
	sqlStatement := `DELETE FROM patient
  WHERE patient.patientid = $1
  `
	_, err := p.db.Exec(sqlStatement, id)
	return err
}

func (p Patient) Update(patient models.Patient) (models.Patient, error) {
	sqlStatement := `UPDATE patient
SET username = $2, full_name = $3, email = $4,dob=$5,contact=$6,bloodgroup=$7,hashed_password=$8,password_changed_at=$9,about=$10,verified=$11,avatar=$12
WHERE patientid = $1
RETURNING patientid,full_name,username,email,dob,contact,bloodgroup;
  `
	var user models.Patient
	err := p.db.QueryRow(sqlStatement, patient.Patientid, patient.Username, patient.Full_name, patient.Email, patient.Dob, patient.Contact, patient.Bloodgroup, patient.Hashed_password, patient.Password_change_at, patient.About, patient.Verified, patient.Avatar).Scan(
		&user.Patientid,
		&user.Full_name,
		&user.Username,
		&user.Email,
		&user.Dob,
		&user.Contact,
		&user.Bloodgroup,
	)
	return user, err
}
