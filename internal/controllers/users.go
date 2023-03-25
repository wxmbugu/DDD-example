package controllers

import (
	"context"
	"database/sql"
	"github.com/patienttracker/internal/models"
	"log"
)

type Users struct {
	db *sql.DB
}

func (u *Users) Create(users models.Users) (models.Users, error) {
	sqlStatement := `
  INSERT INTO users (email,password,roleid) 
  VALUES($1,$2,$3)
  RETURNING *
  `
	err := u.db.QueryRow(sqlStatement, users.Email, users.Password, users.Roleid).Scan(
		&users.Id,
		&users.Email,
		&users.Password,
		&users.Roleid,
	)
	return users, err

}

func (u *Users) Find(id int) (models.Users, error) {
	sqlStatement := `
  SELECT * FROM users
  WHERE users.id = $1
  `
	var user models.Users
	err := u.db.QueryRowContext(context.Background(), sqlStatement, id).Scan(
		&user.Id,
		&user.Email,
		&user.Password,
		&user.Roleid,
	)
	return user, err
}

func (u *Users) FindbyEmail(email string) (models.Users, error) {
	sqlStatement := `
  SELECT * FROM users
  WHERE users.email = $1
  `
	var user models.Users
	err := u.db.QueryRowContext(context.Background(), sqlStatement, email).Scan(
		&user.Id,
		&user.Email,
		&user.Password,
		&user.Roleid,
	)
	return user, err
}
func (u *Users) FindbyRoleId(id int) ([]models.Users, error) {
	sqlStatement := `
 SELECT * FROM users
 WHERE users.roleid = $1
 ORDER BY id
  `
	rows, err := u.db.Query(sqlStatement, id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var items []models.Users
	for rows.Next() {
		var user models.Users
		if err := rows.Scan(
			&user.Id,
			&user.Email,
			&user.Password,
			&user.Roleid,
		); err != nil {
			return nil, err
		}
		items = append(items, user)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil

}

func (u *Users) FindAll() ([]models.Users, error) {
	sqlStatement := `
 SELECT * FROM users
 ORDER BY id 
  `
	rows, err := u.db.QueryContext(context.Background(), sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var items []models.Users
	for rows.Next() {
		var user models.Users
		if err := rows.Scan(
			&user.Id,
			&user.Email,
			&user.Password,
			&user.Roleid,
		); err != nil {
			return nil, err
		}
		items = append(items, user)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (u *Users) Delete(id int) error {
	sqlStatement := `DELETE FROM users 
  WHERE id  = $1
  `
	_, err := u.db.Exec(sqlStatement, id)
	return err
}

func (u *Users) Update(users models.Users) (models.Users, error) {
	sqlStatement := `UPDATE users
SET email=$2,password=$3,roleid=$4
WHERE users.id = $1
RETURNING *;
  `
	var user models.Users
	err := u.db.QueryRow(sqlStatement, users.Id, users.Email, users.Password, users.Roleid).Scan(
		&user.Id,
		&user.Email,
		&user.Password,
		&user.Roleid,
	)
	return user, err
}
