package controllers

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/patienttracker/internal/models"
)

type Appointment struct {
	db *sql.DB
}

func (a Appointment) Create(appointment models.Appointment) (models.Appointment, error) {
	sqlStatement := `
  INSERT INTO appointment (appointmentdate,doctorid,patientid,duration,approval) 
  VALUES ($1,$2,$3,$4,$5)
  RETURNING *
  `
	err := a.db.QueryRow(sqlStatement, appointment.Appointmentdate, appointment.Doctorid, appointment.Patientid, appointment.Duration, appointment.Approval).Scan(
		&appointment.Appointmentid,
		&appointment.Doctorid,
		&appointment.Patientid,
		&appointment.Appointmentdate,
		&appointment.Duration,
		&appointment.Approval)
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
		&appointment.Duration,
		&appointment.Approval,
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
			&i.Duration,
			&i.Approval,
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

func (a Appointment) FindAllByDoctor(id int) ([]models.Appointment, error) {
	sqlStatement := `
	SELECT * FROM appointment 
	WHERE appointment.doctorid = $1
	ORDER BY appointmentid
	LIMIT 15
  `
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := a.db.PrepareContext(ctx, sqlStatement)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, id)
	if err != nil {
		return nil, err
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
			&i.Duration,
			&i.Approval,
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

func (a Appointment) FindAllByPatient(id int) ([]models.Appointment, error) {
	sqlStatement := `
	SELECT * FROM appointment 
	WHERE appointment.patientid = $1
	ORDER BY appointmentid
	LIMIT 15
  `
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := a.db.PrepareContext(ctx, sqlStatement)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, id)
	if err != nil {
		return nil, err
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
			&i.Duration,
			&i.Approval,
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

func (p Appointment) Update(update models.AppointmentUpdate) (models.AppointmentUpdate, error) {
	sqlStatement := `UPDATE appointment
SET appointmentdate = $2,duration = $3,approval = $4
WHERE appointmentid = $1
RETURNING appointmentid,appointmentdate,duration,approval;
  `
	var appointment models.AppointmentUpdate
	err := p.db.QueryRow(sqlStatement, update.Appointmentid, update.Appointmentdate, update.Duration, update.Approval).Scan(
		&appointment.Appointmentid,
		&appointment.Appointmentdate,
		&appointment.Duration,
		&appointment.Approval,
	)
	if err != nil {
		log.Fatal(err)
	}
	return appointment, nil
}
