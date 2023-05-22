package controllers

import (
	"context"
	"database/sql"
	"log"

	"github.com/patienttracker/internal/models"
)

type Department struct {
	db *sql.DB
}

func (d Department) Create(dept models.Department) (models.Department, error) {
	sqlStatement := `
  INSERT INTO department (departmentname) 
  VALUES($1)
  RETURNING *
  `
	var department models.Department
	err := d.db.QueryRow(sqlStatement, dept.Departmentname).Scan(
		&department.Departmentid,
		&department.Departmentname,
	)
	return department, err

}

func (d Department) Find(id int) (models.Department, error) {
	sqlStatement := `
  SELECT * FROM department
  WHERE department.departmentid = $1
  `
	var department models.Department
	err := d.db.QueryRowContext(context.Background(), sqlStatement, id).Scan(
		&department.Departmentid,
		&department.Departmentname,
	)
	return department, err
}

func (d Department) FindbyName(name string) (models.Department, error) {
	sqlStatement := `
	SELECT * FROM department
	WHERE department.departmentname = $1
  `
	var department models.Department
	err := d.db.QueryRowContext(context.Background(), sqlStatement, name).Scan(
		&department.Departmentid,
		&department.Departmentname,
	)
	return department, err
}

func (d Department) FindAll(args models.Filters) ([]models.Department, *models.Metadata, error) {
	var count = 0
	var metadata models.Metadata
	sqlStatement := `
 SELECT  count(*) OVER(), * FROM department
 ORDER BY departmentid
 LIMIT $1
 OFFSET $2
  `
	rows, err := d.db.QueryContext(context.Background(), sqlStatement, args.Limit(), args.Offset())
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var items []models.Department
	for rows.Next() {
		var i models.Department
		if err := rows.Scan(
			&count,
			&i.Departmentid,
			&i.Departmentname,
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

func (d Department) Delete(id int) error {
	sqlStatement := `DELETE FROM department
  WHERE departmentid  = $1
  `
	_, err := d.db.Exec(sqlStatement, id)
	return err
}

func (d Department) Update(update models.Department) (models.Department, error) {
	sqlStatement := `UPDATE department
SET departmentname = $2
WHERE departmentid = $1
RETURNING *;
  `
	var department models.Department
	err := d.db.QueryRow(sqlStatement, update.Departmentid, update.Departmentname).Scan(
		&department.Departmentid,
		&department.Departmentname,
	)
	return department, err
}
