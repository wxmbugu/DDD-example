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
  INSERT INTO schedule (doctorid,type,starttime,endtime,active) 
  VALUES($1,$2,$3,$4,$5)
  RETURNING *
  `
	err := s.db.QueryRow(sqlStatement, schedule.Doctorid, schedule.Type,
		schedule.Starttime, schedule.Endtime, schedule.Active).Scan(
		&schedule.Scheduleid,
		&schedule.Doctorid,
		&schedule.Type,
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
		&schedule.Type,
		&schedule.Starttime,
		&schedule.Endtime,
		&schedule.Active,
	)
	return schedule, err
}
func (s Schedule) FindbyDoctor(id int) ([]models.Schedule, error) {
	sqlStatement := `
 SELECT scheduleid,doctorid,type,starttime,endtime,active FROM schedule
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
			&schedule.Type,
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

type ListSchedule struct {
	Limit  int
	Offset int
}

func (s Schedule) FindAll() ([]models.Schedule, error) {
	sqlStatement := `
 SELECT scheduleid,doctorid,type,starttime,endtime,active FROM schedule
 ORDER BY scheduleid
 LIMIT $1
  `
	rows, err := s.db.QueryContext(context.Background(), sqlStatement, 1)
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
			&schedule.Type,
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

func (s Schedule) Delete(id int) error {
	sqlStatement := `DELETE FROM schedule 
  WHERE scheduleid  = $1
  `
	_, err := s.db.Exec(sqlStatement, id)
	return err
}

func (s Schedule) Update(schedule models.UpdateSchedule, id int) (models.Schedule, error) {
	sqlStatement := `UPDATE schedule
SET type = $2,starttime = $3,endtime=$4,active=$5
WHERE scheduleid = $1
RETURNING *;
  `
	var sched models.Schedule
	err := s.db.QueryRow(sqlStatement, id, schedule.Type,
		schedule.Starttime, schedule.Endtime, schedule.Active).Scan(
		&sched.Scheduleid,
		&sched.Doctorid,
		&sched.Type,
		&sched.Starttime,
		&sched.Endtime,
		&sched.Active,
	)
	return sched, err
}
