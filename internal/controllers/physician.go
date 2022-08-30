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
  INSERT INTO physician (username,hashed_password,full_name,email,contact,departmentname) 
  VALUES($1,$2,$3,$4,$5,$6)
  RETURNING *
  `
	err := p.db.QueryRow(sqlStatement, physician.Username, physician.Hashed_password,
		physician.Full_name, physician.Email, physician.Contact, physician.Departmentname).Scan(
		&physician.Physicianid,
		&physician.Username,
		&physician.Hashed_password,
		&physician.Full_name,
		&physician.Email,
		&physician.Password_changed_at,
		&physician.Created_at,
		&physician.Contact,
		&physician.Departmentname,
	)
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
		&doc.Full_name,
		&doc.Email,
		&doc.Password_changed_at,
		&doc.Created_at,
		&doc.Contact,
		&doc.Departmentname,
	)
	return doc, err
}

func (p Physician) FindDoctorsbyDept(args models.ListDoctorsbyDeptarment) ([]models.Physician, error) {
	sqlStatement := `
	SELECT doctorid, username,full_name,email,created_at,contact,departmentname FROM physician
	WHERE departmentname = $1
	ORDER BY doctorid
	LIMIT $2
	OFFSET $3
  `
	rows, err := p.db.QueryContext(context.Background(), sqlStatement, args.Department, args.Limit, args.Offset)
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
			&i.Contact,
			&i.Departmentname,
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

func (p Physician) FindAll(args models.ListDoctors) ([]models.Physician, error) {
	sqlStatement := `
 SELECT doctorid, username,full_name,email,created_at,contact,departmentname FROM physician
 ORDER BY doctorid
 LIMIT $1
 OFFSET $2
  `
	rows, err := p.db.QueryContext(context.Background(), sqlStatement, args.Limit, args.Offset)
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
			&i.Contact,
			&i.Departmentname,
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

func (p Physician) Update(doctor models.Physician) (models.Physician, error) {
	sqlStatement := `UPDATE physician
SET username = $2, full_name = $3, email = $4,hashed_password=$5,password_changed_at=$6,contact = $7,departmentname=$8
WHERE doctorid = $1
RETURNING doctorid,full_name,username,email,contact,departmentname;
  `
	var doc models.Physician
	err := p.db.QueryRow(sqlStatement, doctor.Physicianid, doctor.Username, doctor.Full_name, doctor.Email, doctor.Hashed_password, doctor.Password_changed_at, doctor.Contact, doctor.Departmentname).Scan(
		&doc.Physicianid,
		&doc.Full_name,
		&doc.Username,
		&doc.Email,
		&doc.Contact,
		&doc.Departmentname,
	)
	if err != nil {
		log.Fatal(err)
	}
	return doc, nil
}
