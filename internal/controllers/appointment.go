package controllers

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/patienttracker/internal/models"
	"time"
)

type Appointment struct {
	db *sql.DB
}

func (a *Appointment) Create(appointment models.Appointment) (models.Appointment, error) {
	sqlStatement := `
  INSERT INTO appointment (appointmentdate,doctorid,patientid,duration,approval,outbound) 
  VALUES ($1,$2,$3,$4,$5,$6)
  RETURNING *
  `
	err := a.db.QueryRow(sqlStatement, appointment.Appointmentdate, appointment.Doctorid, appointment.Patientid, appointment.Duration, appointment.Approval, appointment.Outbound).Scan(
		&appointment.Appointmentid,
		&appointment.Doctorid,
		&appointment.Patientid,
		&appointment.Appointmentdate,
		&appointment.Duration,
		&appointment.Approval,
		&appointment.Outbound)
	return appointment, err

}
func (a *Appointment) Find(id int) (models.Appointment, error) {
	sqlStatement := `
  SELECT * FROM appointment
  WHERE appointment.appointmentid = $1 LIMIT 1
  `

	var appointment models.Appointment
	err := a.db.QueryRowContext(context.Background(), sqlStatement, id).Scan(
		&appointment.Appointmentid,
		&appointment.Doctorid,
		&appointment.Patientid,
		&appointment.Appointmentdate,
		&appointment.Duration,
		&appointment.Approval,
		&appointment.Outbound)
	return appointment, err
}

func (a *Appointment) FindAll(args models.Filters) ([]models.Appointment, *models.Metadata, error) {
	var items []models.Appointment
	var count = 0
	var metadata models.Metadata
	sqlStatement := `
	SELECT count(*) OVER(),* FROM appointment 
	ORDER BY appointmentid
	LIMIT $1
	OFFSET $2
  `
	rows, err := a.db.QueryContext(context.Background(), sqlStatement, args.Limit(), args.Offset())
	if err != nil {
		return items, &metadata, err
	}
	defer rows.Close()
	for rows.Next() {
		var i models.Appointment
		if err := rows.Scan(
			&count,
			&i.Appointmentid,
			&i.Doctorid,
			&i.Patientid,
			&i.Appointmentdate,
			&i.Duration,
			&i.Approval,
			&i.Outbound); err != nil {
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

func (a *Appointment) FindAllByDoctor(id int) ([]models.Appointment, error) {
	sqlStatement := `
	SELECT * FROM appointment 
	WHERE appointment.doctorid = $1
	ORDER BY appointmentid
  `
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := a.db.PrepareContext(ctx, sqlStatement)
	if err != nil {
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
			&i.Outbound); err != nil {
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

func (a *Appointment) FindAllByPatient(id int) ([]models.Appointment, error) {
	sqlStatement := `
	SELECT * FROM appointment 
	WHERE appointment.patientid = $1
	ORDER BY appointmentid
  `
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := a.db.PrepareContext(ctx, sqlStatement)
	if err != nil {
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
			&i.Outbound); err != nil {
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

func (a *Appointment) Delete(id int) error {
	sqlStatement := `DELETE FROM appointment
  WHERE appointment.appointmentid = $1
  `
	_, err := a.db.Exec(sqlStatement, id)
	return err
}

func (p *Appointment) Update(update models.Appointment) (models.Appointment, error) {
	sqlStatement := `UPDATE appointment
SET appointmentdate = $2,duration = $3,approval = $4,outbound = $5
WHERE appointmentid = $1
RETURNING *;
  `
	var appointment models.Appointment
	err := p.db.QueryRow(sqlStatement, update.Appointmentid, update.Appointmentdate, update.Duration, update.Approval, update.Outbound).Scan(
		&appointment.Appointmentid,
		&appointment.Doctorid,
		&appointment.Patientid,
		&appointment.Appointmentdate,
		&appointment.Duration,
		&appointment.Approval,
		&appointment.Outbound)
	if err != nil {
		return appointment, err
	}
	return appointment, nil
}
