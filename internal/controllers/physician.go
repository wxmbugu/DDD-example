package controllers

import (
	"context"
	"database/sql"
	"time"

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
  INSERT INTO physician (username,hashed_password,full_name,email,contact,departmentname,about,verified,avatar) 
  VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)
  RETURNING *
  `
	err := p.db.QueryRow(sqlStatement, physician.Username, physician.Hashed_password,
		physician.Full_name, physician.Email, physician.Contact, physician.Departmentname, physician.About, physician.Verified, physician.Avatar).Scan(
		&physician.Physicianid,
		&physician.Username,
		&physician.Hashed_password,
		&physician.Full_name,
		&physician.Email,
		&physician.About,
		&physician.Avatar,
		&physician.Verified,
		&physician.Password_changed_at,
		&physician.Created_at,
		&physician.Contact,
		&physician.Departmentname,
	)
	if err != nil {
		return models.Physician{}, err
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
		&doc.About,
		&doc.Avatar,
		&doc.Verified,
		&doc.Password_changed_at,
		&doc.Created_at,
		&doc.Contact,
		&doc.Departmentname,
	)
	return doc, err
}

func (p Physician) FindbyEmail(email string) (models.Physician, error) {
	sqlStatement := `
  SELECT * FROM physician
  WHERE physician.email = $1
  `
	var doc models.Physician
	err := p.db.QueryRowContext(context.Background(), sqlStatement, email).Scan(
		&doc.Physicianid,
		&doc.Username,
		&doc.Hashed_password,
		&doc.Full_name,
		&doc.Email,
		&doc.About,
		&doc.Avatar,
		&doc.Verified,
		&doc.Password_changed_at,
		&doc.Created_at,
		&doc.Contact,
		&doc.Departmentname,
	)
	return doc, err
}

func (p Physician) Count() (int, error) {

	counter := 0
	rows, err := p.db.Query("SELECT * FROM physician")
	if err != nil {
		return counter, err
	}
	defer rows.Close()

	for rows.Next() {
		counter++
	}
	return counter, nil
}

func (p Physician) FindDoctorsbyDept(args models.ListDoctorsbyDeptarment) ([]models.Physician, error) {
	sqlStatement := `
	SELECT doctorid, username,full_name,email,about,created_at,contact,departmentname FROM physician
	WHERE departmentname = $1
	ORDER BY doctorid
	LIMIT $2
	OFFSET $3
  `
	rows, err := p.db.QueryContext(context.Background(), sqlStatement, args.Department, args.Limit, args.Offset)
	if err != nil {
		return []models.Physician{}, err
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
			&i.About,
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
		return []models.Physician{}, err
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
SET username = $2, full_name = $3, email = $4,hashed_password=$5,password_changed_at=$6,contact = $7,departmentname=$8,about = $9,verified = $10,avatar = $11
WHERE doctorid = $1
RETURNING doctorid,full_name,username,email,contact,departmentname;
  `
	var doc models.Physician
	err := p.db.QueryRow(sqlStatement, doctor.Physicianid, doctor.Username, doctor.Full_name, doctor.Email, doctor.Hashed_password, doctor.Password_changed_at, doctor.Contact, doctor.Departmentname, doctor.About, doctor.Verified, doctor.Avatar).Scan(
		&doc.Physicianid,
		&doc.Full_name,
		&doc.Username,
		&doc.Email,
		&doc.Contact,
		&doc.Departmentname,
	)
	if err != nil {
		return doc, err
	}
	return doc, nil
}

func (p Physician) Filter(username string, departmentname string, filters models.Filters) ([]*models.Physician, *models.Metadata, error) {
	var metadata models.Metadata
	counter := 0
	query := `
SELECT count(*) OVER(),doctorid,full_name,username,email,contact,departmentname 
FROM physician
WHERE (username ILIKE '%' || $1 || '%' OR $1 = '')
AND (departmentname ILIKE '%' || $2 || '%' OR $2 = '')
ORDER BY doctorid ASC LIMIT $3 OFFSET $4`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Pass the title and genres as the placeholder parameter values.
	rows, err := p.db.QueryContext(ctx, query, username, departmentname, filters.Limit(), filters.Offset())
	if err != nil {
		return nil, &metadata, err
	}
	defer rows.Close()
	doctors := []*models.Physician{}
	for rows.Next() {
		var doctor models.Physician
		err := rows.Scan(
			&counter,
			&doctor.Physicianid,
			&doctor.Full_name,
			&doctor.Username,
			&doctor.Email,
			&doctor.Contact,
			&doctor.Departmentname,
		)
		if err != nil {
			return nil, &metadata, err
		}
		doctors = append(doctors, &doctor)
	}
	if err = rows.Err(); err != nil {
		return nil, &metadata, err
	}
	metadata = models.CalculateMetadata(counter, filters.Page, filters.PageSize)
	return doctors, &metadata, nil
}
