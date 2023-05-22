package controllers

import (
	"context"
	"database/sql"
	"log"

	"github.com/patienttracker/internal/models"
)

type Schedule struct {
	db *sql.DB
}

func (s Schedule) Create(schedule models.Schedule) (models.Schedule, error) {
	sqlStatement := `
  INSERT INTO schedule (doctorid,starttime,endtime,active) 
  VALUES($1,$2,$3,$4)
  RETURNING *
  `
	err := s.db.QueryRow(sqlStatement, schedule.Doctorid, schedule.Starttime, schedule.Endtime, schedule.Active).Scan(
		&schedule.Scheduleid,
		&schedule.Doctorid,
		&schedule.Starttime,
		&schedule.Endtime,
		&schedule.Active,
	)
	return schedule, err

}

func (s Schedule) Find(id int) (models.Schedule, error) {
	sqlStatement := `
  SELECT * FROM schedule
  WHERE schedule.scheduleid = $1
  `
	var schedule models.Schedule
	err := s.db.QueryRowContext(context.Background(), sqlStatement, id).Scan(
		&schedule.Scheduleid,
		&schedule.Doctorid,
		&schedule.Starttime,
		&schedule.Endtime,
		&schedule.Active,
	)
	return schedule, err
}
func (s Schedule) FindbyDoctor(id int) ([]models.Schedule, error) {
	sqlStatement := `
 SELECT scheduleid,doctorid,starttime,endtime,active FROM schedule
 WHERE schedule.doctorid = $1
 ORDER BY scheduleid
  `
	rows, err := s.db.Query(sqlStatement, id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var items []models.Schedule
	for rows.Next() {
		var schedule models.Schedule
		if err := rows.Scan(
			&schedule.Scheduleid,
			&schedule.Doctorid,
			&schedule.Starttime,
			&schedule.Endtime,
			&schedule.Active,
		); err != nil {
			return nil, err
		}
		items = append(items, schedule)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil

}
func (s Schedule) FindAll(args models.Filters) ([]models.Schedule, *models.Metadata, error) {
	var count = 0
	var metadata models.Metadata
	sqlStatement := `
 SELECT  count(*) OVER(),scheduleid,doctorid,starttime,endtime,active FROM schedule
 ORDER BY scheduleid
 LIMIT $1
 OFFSET $2
  `
	rows, err := s.db.QueryContext(context.Background(), sqlStatement, args.Limit(), args.Offset())
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var items []models.Schedule
	for rows.Next() {
		var schedule models.Schedule
		if err := rows.Scan(
			&count,
			&schedule.Scheduleid,
			&schedule.Doctorid,
			&schedule.Starttime,
			&schedule.Endtime,
			&schedule.Active,
		); err != nil {
			return nil, &metadata, err
		}
		items = append(items, schedule)
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

func (s Schedule) Delete(id int) error {
	sqlStatement := `DELETE FROM schedule 
  WHERE scheduleid  = $1
  `
	_, err := s.db.Exec(sqlStatement, id)
	return err
}

func (s Schedule) Update(schedule models.Schedule) (models.Schedule, error) {
	sqlStatement := `UPDATE schedule
SET starttime = $2,endtime=$3,active=$4
WHERE scheduleid = $1
RETURNING *;
  `
	var sched models.Schedule
	err := s.db.QueryRow(sqlStatement, schedule.Scheduleid, schedule.Starttime, schedule.Endtime, schedule.Active).Scan(
		&sched.Scheduleid,
		&sched.Doctorid,
		&sched.Starttime,
		&sched.Endtime,
		&sched.Active,
	)
	return sched, err
}
