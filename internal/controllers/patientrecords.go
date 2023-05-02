package controllers

import (
	"context"
	"database/sql"
	"github.com/patienttracker/internal/models"
)

type PatientRecords struct {
	db *sql.DB
}

func (p PatientRecords) Create(patientrecords models.Patientrecords) (models.Patientrecords, error) {
	sqlStatement := `
  INSERT INTO patientrecords (patientid,date,height,bloodpressure,heartrate,temperature,weight,doctorid,additional,nurseid) 
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
  RETURNING *
  `
	err := p.db.QueryRow(sqlStatement, patientrecords.Patienid, patientrecords.Date,
		patientrecords.Height, patientrecords.Bp, patientrecords.HeartRate, patientrecords.Temperature, patientrecords.Weight, patientrecords.Doctorid, patientrecords.Additional, patientrecords.Nurseid).Scan(
		&patientrecords.Recordid,
		&patientrecords.Patienid,
		&patientrecords.Date,
		&patientrecords.Height,
		&patientrecords.Bp,
		&patientrecords.HeartRate,
		&patientrecords.Temperature,
		&patientrecords.Weight,
		&patientrecords.Doctorid,
		&patientrecords.Additional,
		&patientrecords.Nurseid)
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
		&record.Height,
		&record.Bp,
		&record.HeartRate,
		&record.Temperature,
		&record.Weight,
		&record.Doctorid,
		&record.Additional,
		&record.Nurseid)
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
		return []models.Patientrecords{}, err
	}
	defer rows.Close()
	var items []models.Patientrecords
	for rows.Next() {
		var record models.Patientrecords
		if err := rows.Scan(
			&record.Recordid,
			&record.Patienid,
			&record.Date,
			&record.Height,
			&record.Bp,
			&record.HeartRate,
			&record.Temperature,
			&record.Weight,
			&record.Doctorid,
			&record.Additional,
			&record.Nurseid); err != nil {
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

func (p PatientRecords) Count() (int, error) {

	counter := 0
	rows, err := p.db.Query("SELECT * FROM patientrecords")
	if err != nil {
		return counter, err
	}
	defer rows.Close()

	for rows.Next() {
		// you can even scan+store the result if you need them later
		counter++
	}
	return counter, nil
}

func (p PatientRecords) FindAllByPatient(id int) ([]models.Patientrecords, error) {
	sqlStatement := `
SELECT * FROM patientrecords
WHERE patientid = $1
 ORDER BY recordid
 
  `
	rows, err := p.db.QueryContext(context.Background(), sqlStatement, id)
	if err != nil {
		return []models.Patientrecords{}, err
	}
	defer rows.Close()
	var items []models.Patientrecords
	for rows.Next() {
		var record models.Patientrecords
		if err := rows.Scan(
			&record.Recordid,
			&record.Patienid,
			&record.Date,
			&record.Height,
			&record.Bp,
			&record.HeartRate,
			&record.Temperature,
			&record.Weight,
			&record.Doctorid,
			&record.Additional,
			&record.Nurseid); err != nil {
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
		return []models.Patientrecords{}, err
	}
	defer rows.Close()
	var items []models.Patientrecords
	for rows.Next() {
		var record models.Patientrecords
		if err := rows.Scan(
			&record.Recordid,
			&record.Patienid,
			&record.Date,
			&record.Height,
			&record.Bp,
			&record.HeartRate,
			&record.Temperature,
			&record.Weight,
			&record.Doctorid,
			&record.Additional,
			&record.Nurseid); err != nil {
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
SET height = $2, bloodpressure = $3, temperature = $4,weight=$5,additional=$6
WHERE recordid = $1
RETURNING *;
  `
	var precord models.Patientrecords
	err := p.db.QueryRow(sqlStatement, record.Recordid, record.Height, record.Bp, record.Temperature, record.Weight, record.Additional).Scan(
		&precord.Recordid,
		&precord.Patienid,
		&precord.Date,
		&precord.Height,
		&precord.Bp,
		&precord.HeartRate,
		&precord.Temperature,
		&precord.Weight,
		&precord.Doctorid,
		&precord.Additional,
		&precord.Nurseid)
	return precord, err
}
