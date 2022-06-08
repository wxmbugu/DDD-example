package controllers

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/patienttracker/internal/models"
)

type Appointment struct {
	db *sql.DB
}

func (a Appointment) Create(appointment models.Appointment) (models.Appointment, error) {
	sqlStatement := `
  INSERT INTO appointment (appointmentdate,doctorid,patientid) 
  VALUES ($1,$2,$3)
  RETURNING *
  `
	err := a.db.QueryRow(sqlStatement, appointment.Appointmentdate, appointment.Doctorid, appointment.Patientid).Scan(
		&appointment.Appointmentid,
		&appointment.Patientid,
		&appointment.Doctorid,
		&appointment.Appointmentdate)
	return appointment, err

}

func (a Appointment) Find(id int) (models.Appointment, error) {
	sqlStatement := `
  SELECT * FROM appointment
  WHERE appointment.appointmentid = $1 LIMIT 1
  `
	var appointment models.Appointment
	err := a.db.QueryRowContext(context.Background(), sqlStatement, id).Scan(
		&appointment.Appointmentid,
		&appointment.Patientid,
		&appointment.Doctorid,
		&appointment.Appointmentdate,
	)
	return appointment, err
}

type ListAppointment struct {
	Limit  int
	Offset int
}

func (a Appointment) FindAll() ([]models.Appointment, error) {
	sqlStatement := `
	SELECT * FROM appointment 
	ORDER BY appointmentid
	LIMIT $1
  `
	rows, err := a.db.QueryContext(context.Background(), sqlStatement, 10)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var items []models.Appointment
	for rows.Next() {
		var i models.Appointment
		if err := rows.Scan(
			&i.Appointmentid,
			&i.Doctorid,
			&i.Patientid,
			&i.Appointmentdate,
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

func (a Appointment) Delete(id int) error {
	sqlStatement := `DELETE FROM appointment
  WHERE appointment.appointmentid = $1
  `
	_, err := a.db.Exec(sqlStatement, id)
	return err
}

func (p Appointment) Update(date time.Time, id int) (time.Time, error) {
	sqlStatement := `UPDATE appointment
SET appointmentdate = $2
WHERE appointmentid = $1
RETURNING appointmentdate;
  `
	var appointment models.Appointment
	err := p.db.QueryRow(sqlStatement, id, date).Scan(
		&appointment.Appointmentdate,
	)
	if err != nil {
		log.Fatal(err)
	}
	return appointment.Appointmentdate, nil
}
