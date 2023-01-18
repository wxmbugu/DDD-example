package controllers

import (
	"context"
	"database/sql"
	"github.com/patienttracker/internal/models"
	"log"
)

type Permissions struct {
	db *sql.DB
}

func (p *Permissions) Create(perm models.Permissions) (models.Permissions, error) {
	sqlStatement := `
  INSERT INTO permissions (permission,roleid) 
  VALUES($1,$2)
  RETURNING *
  `
	err := p.db.QueryRow(sqlStatement, perm.Permission, perm.Roleid).Scan(
		&perm.Permissionid,
		&perm.Permission,
		&perm.Roleid,
	)
	return perm, err

}

func (p *Permissions) Find(id int) (models.Permissions, error) {
	sqlStatement := `
  SELECT * FROM permissions
  WHERE permissions.permissionid = $1
  `
	var perm models.Permissions
	err := p.db.QueryRowContext(context.Background(), sqlStatement, id).Scan(
		&perm.Permissionid,
		&perm.Permission,
		&perm.Roleid,
	)
	return perm, err
}
func (p *Permissions) FindbyRoleId(id int) ([]models.Permissions, error) {
	sqlStatement := `
 SELECT * FROM permissions
 WHERE permissions.roleid = $1
 ORDER BY permissionid
  `
	rows, err := p.db.Query(sqlStatement, id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var items []models.Permissions
	for rows.Next() {
		var perm models.Permissions
		if err := rows.Scan(
			&perm.Permissionid,
			&perm.Permission,
			&perm.Roleid,
		); err != nil {
			return nil, err
		}
		items = append(items, perm)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil

}

func (p *Permissions) FindAll() ([]models.Permissions, error) {
	sqlStatement := `
 SELECT * FROM permissions
 ORDER BY permissionid
  `
	rows, err := p.db.QueryContext(context.Background(), sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var items []models.Permissions
	for rows.Next() {
		var perm models.Permissions
		if err := rows.Scan(
			&perm.Permissionid,
			&perm.Permission,
			&perm.Roleid,
		); err != nil {
			return nil, err
		}
		items = append(items, perm)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (p *Permissions) Delete(id int) error {
	sqlStatement := `DELETE FROM permissions 
  WHERE permissionid  = $1
  `
	_, err := p.db.Exec(sqlStatement, id)
	return err
}

func (p *Permissions) Update(perms models.Permissions) (models.Permissions, error) {
	sqlStatement := `UPDATE permissions
SET permission=$2,roleid=$3
WHERE permissionid = $1
RETURNING *;
  `
	var perm models.Permissions
	err := p.db.QueryRow(sqlStatement, perms.Permissionid, perms.Permission, perms.Roleid).Scan(
		&perm.Permissionid,
		&perm.Permission,
		&perm.Roleid,
	)
	return perm, err
}
