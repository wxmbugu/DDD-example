package controllers

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/patienttracker/internal/models"
)

type Patient struct {
	db *sql.DB
}

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

func (p Patient) FindAll(args models.Filters) ([]models.Patient, *models.Metadata, error) {
	var count = 0
	var metadata models.Metadata
	sqlStatement := `
 SELECT  count(*) OVER(),patientid, username,full_name,email,dob,contact,bloodgroup,created_at,ischild FROM patient
 ORDER BY patientid
 LIMIT $1
 OFFSET $2
  `
	rows, err := p.db.QueryContext(context.Background(), sqlStatement, args.Limit(), args.Offset())
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var items []models.Patient
	for rows.Next() {
		var i models.Patient
		if err := rows.Scan(
			&count,
			&i.Patientid,
			&i.Username,
			&i.Full_name,
			&i.Email,
			&i.Dob,
			&i.Contact,
			&i.Bloodgroup,
			&i.Created_at,
			&i.Ischild); err != nil {
			return nil, &metadata, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, &metadata, err
	}
	if err := rows.Err(); err != nil {
		return nil, &metadata, err
	}
	metadata = models.CalculateMetadata(count, args.Page, args.PageSize)
	return items, &metadata, nil
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
SET username = $2, full_name = $3, email = $4,dob=$5,contact=$6,bloodgroup=$7,hashed_password=$8,password_changed_at=$9,about=$10,verified=$11,avatar=$12,ischild=$13
WHERE patientid = $1
RETURNING patientid,full_name,username,email,dob,contact,bloodgroup,ischild;
  `
	var user models.Patient
	err := p.db.QueryRow(sqlStatement, patient.Patientid, patient.Username, patient.Full_name, patient.Email, patient.Dob, patient.Contact, patient.Bloodgroup, patient.Hashed_password, patient.Password_change_at, patient.About, patient.Verified, patient.Avatar, patient.Ischild).Scan(
		&user.Patientid,
		&user.Full_name,
		&user.Username,
		&user.Email,
		&user.Dob,
		&user.Contact,
		&user.Bloodgroup,
		&user.Ischild,
	)
	return user, err
}
func (p Patient) Filter(username string, filters models.Filters) ([]*models.Patient, *models.Metadata, error) {
	var metadata models.Metadata
	counter := 0
	query := `
SELECT count(*) OVER(),patientid,full_name,username,email,dob,contact,bloodgroup,ischild 
FROM patient
WHERE (username ILIKE '%' || $1 || '%' OR $1 = '')
ORDER BY patientid ASC LIMIT $2 OFFSET $3`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Pass the title and genres as the placeholder parameter values.
	rows, err := p.db.QueryContext(ctx, query, username, filters.Limit(), filters.Offset())
	if err != nil {
		return nil, &metadata, err
	}
	defer rows.Close()
	items := []*models.Patient{}
	for rows.Next() {
		var user models.Patient
		err := rows.Scan(
			&counter,
			&user.Patientid,
			&user.Full_name,
			&user.Username,
			&user.Email,
			&user.Dob,
			&user.Contact,
			&user.Bloodgroup,
			&user.Ischild,
		)
		if err != nil {
			return nil, &metadata, err
		}
		items = append(items, &user)
	}
	if err = rows.Err(); err != nil {
		return nil, &metadata, err
	}
	metadata = models.CalculateMetadata(counter, filters.Page, filters.PageSize)
	return items, &metadata, nil
}
