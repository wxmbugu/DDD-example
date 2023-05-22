package controllers

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/patienttracker/internal/models"
)

type Nurse struct {
	db *sql.DB
}

func (n Nurse) Create(nurse models.Nurse) (models.Nurse, error) {
	sqlStatement := `
  INSERT INTO nurse (username,full_name,email,hashed_password) 
  VALUES($1,$2,$3,$4)
  RETURNING *
  `
	err := n.db.QueryRow(sqlStatement, nurse.Username, nurse.Full_name,
		nurse.Email, nurse.Hashed_password).Scan(
		&nurse.Id,
		&nurse.Username,
		&nurse.Full_name,
		&nurse.Email,
		&nurse.Hashed_password,
		&nurse.Password_changed_at,
		&nurse.Created_at,
	)
	if err != nil {
		log.Fatal(err)
	}
	return nurse, nil

}

func (n Nurse) Find(id int) (models.Nurse, error) {
	sqlStatement := `
  SELECT * FROM nurse
  WHERE nurse.id = $1
  `
	var nurse models.Nurse
	err := n.db.QueryRowContext(context.Background(), sqlStatement, id).Scan(
		&nurse.Id,
		&nurse.Username,
		&nurse.Full_name,
		&nurse.Email,
		&nurse.Hashed_password,
		&nurse.Password_changed_at,
		&nurse.Created_at)
	return nurse, err
}

func (n Nurse) FindbyEmail(email string) (models.Nurse, error) {
	sqlStatement := `
  SELECT * FROM nurse
  WHERE nurse.email = $1
  `
	var nurse models.Nurse
	err := n.db.QueryRowContext(context.Background(), sqlStatement, email).Scan(
		&nurse.Id,
		&nurse.Username,
		&nurse.Full_name,
		&nurse.Email,
		&nurse.Hashed_password,
		&nurse.Password_changed_at,
		&nurse.Created_at)
	return nurse, err
}

func (p Nurse) FindAll(args models.Filters) ([]models.Nurse, *models.Metadata, error) {
	var count = 0
	var metadata models.Metadata
	sqlStatement := `
 SELECT  count(*) OVER(),id, username,full_name,email FROM nurse
 ORDER BY id
 LIMIT $1
 OFFSET $2
  `
	rows, err := p.db.QueryContext(context.Background(), sqlStatement, args.Limit(), args.Offset())
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var items []models.Nurse
	for rows.Next() {
		var i models.Nurse
		if err := rows.Scan(
			&count,
			&i.Id,
			&i.Username,
			&i.Full_name,
			&i.Email,
		); err != nil {
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

func (n Nurse) Delete(id int) error {
	sqlStatement := `DELETE FROM nurse
  WHERE id  = $1
  `
	_, err := n.db.Exec(sqlStatement, id)
	return err
}

func (p Nurse) Update(nurse models.Nurse) (models.Nurse, error) {
	sqlStatement := `UPDATE nurse
SET username = $2, full_name = $3, email = $4,hashed_password=$5,password_changed_at=$6
WHERE id = $1
RETURNING id,full_name,username,email;
  `
	var nur models.Nurse
	err := p.db.QueryRow(sqlStatement, nurse.Id, nurse.Username, nurse.Full_name, nurse.Email, nurse.Hashed_password, nurse.Password_changed_at).Scan(
		&nur.Id,
		&nur.Full_name,
		&nur.Username,
		&nur.Email,
	)
	if err != nil {
		return nur, err
	}
	return nur, nil
}

func (p Nurse) Filter(username string, filters models.Filters) ([]*models.Nurse, *models.Metadata, error) {
	var metadata models.Metadata
	counter := 0
	query := `
SELECT count(*) OVER(),id, username,full_name,email 
FROM nurse
WHERE (username ILIKE '%' || $1 || '%' OR $1 = '')
ORDER BY id ASC LIMIT $2 OFFSET $3`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := p.db.QueryContext(ctx, query, username, filters.Limit(), filters.Offset())
	if err != nil {
		return nil, &metadata, err
	}
	defer rows.Close()
	items := []*models.Nurse{}
	for rows.Next() {
		var nurse models.Nurse
		err := rows.Scan(
			&counter,
			&nurse.Id,
			&nurse.Username,
			&nurse.Full_name,
			&nurse.Email,
		)
		if err != nil {
			return nil, &metadata, err
		}
		items = append(items, &nurse)
	}
	if err = rows.Err(); err != nil {
		return nil, &metadata, err
	}
	metadata = models.CalculateMetadata(counter, filters.Page, filters.PageSize)
	return items, &metadata, nil
}
