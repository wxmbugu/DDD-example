package controllers

import (
	"context"
	"database/sql"
	"log"

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

func (p Physician) Create(physician models.Physician) (models.Physician, error) {
	sqlStatement := `
  INSERT INTO physician (username,hashed_password,full_name,email) 
  VALUES($1,$2,$3,$4)
  RETURNING *
  `
	err := p.db.QueryRow(sqlStatement, physician.Username, physician.Hashed_password,
		physician.Full_name, physician.Email).Scan(
		&physician.Physicianid,
		&physician.Username,
		&physician.Hashed_password,
		&physician.Email,
		&physician.Full_name,
		&physician.Password_changed_at,
		&physician.Created_at)
	if err != nil {
		log.Fatal(err)
	}
	return physician, nil

}

func (p Physician) Find(id int) (models.Physician, error) {
	sqlStatement := `
  SELECT * FROM physician
  WHERE physician.doctorid = $1
  `
	var doc models.Physician
	err := p.db.QueryRowContext(context.Background(), sqlStatement, id).Scan(
		&doc.Physicianid,
		&doc.Username,
		&doc.Hashed_password,
		&doc.Email,
		&doc.Full_name,
		&doc.Password_changed_at,
		&doc.Created_at,
	)
	return doc, err
}

type ListPhyysiciant struct {
	Limit  int
	Offset int
}

func (p Physician) FindAll() ([]models.Physician, error) {
	sqlStatement := `
 SELECT doctorid, username,full_name,email,created_at FROM physician
 ORDER BY doctorid
 LIMIT $1
  `
	rows, err := p.db.QueryContext(context.Background(), sqlStatement, 10)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var items []models.Physician
	for rows.Next() {
		var i models.Physician
		if err := rows.Scan(
			&i.Physicianid,
			&i.Username,
			&i.Full_name,
			&i.Email,
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
	sqlStatement := `DELETE FROM physician
  WHERE doctorid  = $1
  `
	_, err := p.db.Exec(sqlStatement, id)
	return err
}

func (p Physician) Update(doctor models.UpdatePhysician, id int) (models.Physician, error) {
	sqlStatement := `UPDATE physician
SET username = $2, full_name = $3, email = $4,hashed_password=$5,password_changed_at=$6
WHERE doctorid = $1
RETURNING doctorid,full_name,username,email;
  `
	var doc models.Physician
	err := p.db.QueryRow(sqlStatement, id, doctor.Username, doctor.Full_name, doctor.Email, doctor.Hashed_password, doctor.Password_changed_at).Scan(
		&doc.Physicianid,
		&doc.Full_name,
		&doc.Username,
		&doc.Email,
	)
	if err != nil {
		log.Fatal(err)
	}
	return doc, nil
}
