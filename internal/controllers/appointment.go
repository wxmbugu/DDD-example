package controllers

import (
	"context"
	"log"
	"time"

	"github.com/patienttracker/internal/db"
	"github.com/patienttracker/internal/models"
)

type Appointment struct {
	db db.Database
}

/*
  appointmentid   int
		doctorid        int
		patientid       int
		appointmentdate time.Time
*/
func NewAppointenttRepositry() models.AppointmentRepository {
	dbconn, err := db.New()
	if err != nil {
		log.Fatal(err)
	}
	return Appointment{
		db: dbconn,
	}
}

func (a Appointment) Create(appointment models.Appointment) (models.Appointment, error) {
	sqlStatement := `
  INSERT INTO appointment (appointmentdate,doctorid,patientid) 
  s.dbconn
  RETURNING *
  `
	err := a.db.Conn.QueryRow(sqlStatement, appointment.Appointmentdate, appointment.Doctorid, appointment.Patientid).Scan(
		&appointment.Appointmentid,
		&appointment.Patientid,
		&appointment.Doctorid,
		&appointment.Appointmentdate)
	if err != nil {
		log.Fatal(err)
	}
	return appointment, nil

}

func (a Appointment) Find(id int) (models.Appointment, error) {
	sqlStatement := `
  SELECT * FROM appointment
  WHERE appointment.appointmentid = $1 LIMIT 1
  `
	var appointment models.Appointment
	err := a.db.Conn.QueryRowContext(context.Background(), sqlStatement, id).Scan(
		&appointment.Appointmentid,
		&appointment.Patientid,
		&appointment.Doctorid,
		&appointment.Appointmentdate,
	)
	if err != nil {
		log.Fatal(err)
	}
	return appointment, nil
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
	rows, err := a.db.Conn.QueryContext(context.Background(), sqlStatement, 10)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var items []models.Appointment
	for rows.Next() {
		var i models.Appointment
		if err := rows.Scan(
			&i.Patientid,
			&i.Appointmentdate,
			&i.Doctorid,
			&i.Appointmentid,
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
	_, err := a.db.Conn.Exec(sqlStatement, id)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (p Appointment) Update(date time.Time, id int) (time.Time, error) {
	sqlStatement := `UPDATE appointment
SET appointment = $2
WHERE appointmentid = $1
RETURNING appointmentdate;
  `
	var appointment models.Appointment
	err := p.db.Conn.QueryRow(sqlStatement, id, date).Scan(
		&appointment.Appointmentdate,
	)
	if err != nil {
		log.Fatal(err)
	}
	return appointment.Appointmentdate, nil
}
