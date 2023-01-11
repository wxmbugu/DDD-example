package controllers

import (
	"context"
	"database/sql"
	"log"

	"github.com/patienttracker/internal/models"
)

type Roles struct {
	db *sql.DB
}

func (r *Roles) Create(roles models.Roles) (models.Roles, error) {
	sqlStatement := `
  INSERT INTO roles (role) 
  VALUES($1)
  RETURNING *
  `
	err := r.db.QueryRow(sqlStatement, roles.Role).Scan(
		&roles.Roleid,
		&roles.Role,
	)
	return roles, err
}

func (r *Roles) Find(id int) (models.Roles, error) {
	sqlStatement := `
  SELECT * FROM roles
  WHERE roles.roleid = $1
  `
	var role models.Roles
	err := r.db.QueryRowContext(context.Background(), sqlStatement, id).Scan(
		&role.Roleid,
		&role.Role,
	)
	return role, err
}

func (r *Roles) FindAll() ([]models.Roles, error) {
	sqlStatement := `
 SELECT * FROM roles
 ORDER BY roleid
  `
	rows, err := r.db.QueryContext(context.Background(), sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var items []models.Roles
	for rows.Next() {
		var role models.Roles
		if err := rows.Scan(
			&role.Roleid,
			&role.Role,
		); err != nil {
			return nil, err
		}
		items = append(items, role)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Roles) Delete(id int) error {
	sqlStatement := `DELETE FROM roles 
  WHERE roleid  = $1
  `
	_, err := r.db.Exec(sqlStatement, id)
	return err
}

func (r *Roles) Update(role models.Roles) (models.Roles, error) {
	sqlStatement := `UPDATE roles
SET role = $2
WHERE roleid = $1
RETURNING *
  `
	var rol models.Roles
	err := r.db.QueryRow(sqlStatement, role.Roleid, role.Role).Scan(
		&rol.Roleid,
		&rol.Role,
	)
	return rol, err
}
