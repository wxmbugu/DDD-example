package controllers

import (
	"context"
	"database/sql"
	"log"

	"github.com/patienttracker/internal/models"
)

type PatientRecords struct {
	db *sql.DB
}

func (p PatientRecords) Create(patientrecords models.Patientrecords) (models.Patientrecords, error) {
	sqlStatement := `
  INSERT INTO patientrecords (patientid,date,disease,prescription,diagnosis,weight,doctorid) 
VALUES ($1,$2,$3,$4,$5,$6,$7)
  RETURNING *
  `
	err := p.db.QueryRow(sqlStatement, patientrecords.Patienid, patientrecords.Date,
		patientrecords.Disease, patientrecords.Prescription, patientrecords.Diagnosis, patientrecords.Weight, patientrecords.Doctorid).Scan(
		&patientrecords.Recordid,
		&patientrecords.Patienid,
		&patientrecords.Date,
		&patientrecords.Disease,
		&patientrecords.Prescription,
		&patientrecords.Diagnosis,
		&patientrecords.Weight,
		&patientrecords.Doctorid)
	return patientrecords, err

}

func (p PatientRecords) Find(id int) (models.Patientrecords, error) {
	sqlStatement := `
	SELECT * FROM patientrecords
  WHERE recordid = $1 LIMIT 1
  `
	var record models.Patientrecords
	err := p.db.QueryRowContext(context.Background(), sqlStatement, id).Scan(
		&record.Recordid,
		&record.Patienid,
		&record.Date,
		&record.Disease,
		&record.Prescription,
		&record.Diagnosis,
		&record.Weight,
		&record.Doctorid,
	)
	return record, err
}

func (p PatientRecords) FindAll(args models.ListPatientRecords) ([]models.Patientrecords, error) {
	sqlStatement := `
SELECT * FROM patientrecords
 ORDER BY recordid
 LIMIT $1
 OFFSET $2
  `
	rows, err := p.db.QueryContext(context.Background(), sqlStatement, args.Limit, args.Offset)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var items []models.Patientrecords
	for rows.Next() {
		var record models.Patientrecords
		if err := rows.Scan(
			&record.Recordid,
			&record.Patienid,
			&record.Date,
			&record.Disease,
			&record.Prescription,
			&record.Diagnosis,
			&record.Weight,
			&record.Doctorid,
		); err != nil {
			return nil, err
		}
		items = append(items, record)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (p PatientRecords) FindAllByPatient(id int) ([]models.Patientrecords, error) {
	sqlStatement := `
SELECT * FROM patientrecords
WHERE patientid = $1
 ORDER BY recordid
 
  `
	rows, err := p.db.QueryContext(context.Background(), sqlStatement, id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var items []models.Patientrecords
	for rows.Next() {
		var record models.Patientrecords
		if err := rows.Scan(
			&record.Recordid,
			&record.Patienid,
			&record.Date,
			&record.Disease,
			&record.Prescription,
			&record.Diagnosis,
			&record.Weight,
			&record.Doctorid,
		); err != nil {
			return nil, err
		}
		items = append(items, record)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (p PatientRecords) FindAllByDoctor(id int) ([]models.Patientrecords, error) {
	sqlStatement := `
SELECT * FROM patientrecords
WHERE doctorid = $1
ORDER BY recordid
  `
	rows, err := p.db.QueryContext(context.Background(), sqlStatement, id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var items []models.Patientrecords
	for rows.Next() {
		var record models.Patientrecords
		if err := rows.Scan(
			&record.Recordid,
			&record.Patienid,
			&record.Date,
			&record.Disease,
			&record.Prescription,
			&record.Diagnosis,
			&record.Weight,
			&record.Doctorid,
		); err != nil {
			return nil, err
		}
		items = append(items, record)
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
	sqlStatement := `DELETE FROM patientrecords
  WHERE recordid = $1
  `
	_, err := p.db.Exec(sqlStatement, id)

	return err
}

func (p PatientRecords) Update(record models.Patientrecords) (models.Patientrecords, error) {
	sqlStatement := `UPDATE patientrecords
SET diagnosis = $2, disease = $3, prescription = $4,weight=$5
WHERE recordid = $1
RETURNING *;
  `
	var precord models.Patientrecords
	err := p.db.QueryRow(sqlStatement, record.Recordid, record.Diagnosis, record.Disease, record.Prescription, record.Weight).Scan(
		&precord.Recordid,
		&precord.Patienid,
		&precord.Date,
		&precord.Disease,
		&precord.Prescription,
		&precord.Diagnosis,
		&precord.Weight,
		&precord.Doctorid,
	)
	return precord, err
}
