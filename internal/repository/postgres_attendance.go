package repository

import (
	"context"
	"database/sql"
	"fmt"

	"ku_backend_new/internal/domain"
)

type PostgresAttendanceRepo struct {
	db *sql.DB
}

// NewPostgresAttendanceRepo creates a new instance of the repository
func NewPostgresAttendanceRepo(db *sql.DB) *PostgresAttendanceRepo {
	return &PostgresAttendanceRepo{db: db}
}

// Save handles a single manual registration
func (r *PostgresAttendanceRepo) Save(ctx context.Context, record *domain.Absence) error {
	// CHANGE 'date' TO 'absence_date' HERE TO MATCH YOUR DATABASE
	query := `INSERT INTO absences (reg_number, first_name, last_name, absence_date, absence_time, group_number, subject, status) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		record.RegNumber,
		record.FirstName,
		record.LastName,
		record.Date,
		record.Time,
		record.Group,
		record.Subject,
		string(record.Status),
	)

	if err != nil {
		fmt.Printf("DATABASE ERROR: %v\n", err)
		return err
	}
	return nil
}

// SaveBatch handles the Excel import by inserting many rows at once
func (r *PostgresAttendanceRepo) SaveBatch(ctx context.Context, records []domain.Absence) error {
	if len(records) == 0 {
		return nil
	}

	// Use a transaction for speed
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO absences (reg_number, first_name, last_name, absence_date, absence_time, group_number, subject, status)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, rec := range records {
		_, err = stmt.ExecContext(ctx, rec.RegNumber, rec.FirstName, rec.LastName, rec.Date, rec.Time, rec.Group, rec.Subject, string(rec.Status))
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (r *PostgresAttendanceRepo) Find(ctx context.Context, group string, regNum string, month int, year int) ([]domain.Absence, error) {
	// Strict month/year filtering
	query := `SELECT reg_number, first_name, last_name, absence_date, absence_time, group_number, subject, status
          FROM absences
          WHERE CAST(EXTRACT(MONTH FROM absence_date) AS INTEGER) = $1
          AND CAST(EXTRACT(YEAR FROM absence_date) AS INTEGER) = $2`

	args := []interface{}{month, year}
	counter := 3

	// Optional group filter
	if group != "" && group != "All" {
		query += fmt.Sprintf(" AND group_number = $%d", counter)
		args = append(args, group)
		counter++
	}

	// Optional search filter (registration number / first / last name)
	if regNum != "" {
		query += fmt.Sprintf(" AND (reg_number ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d)", counter, counter, counter)
		args = append(args, "%"+regNum+"%")
		counter++
	}

	query += " ORDER BY absence_date DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.Absence
	for rows.Next() {
		var a domain.Absence
		if err := rows.Scan(&a.RegNumber, &a.FirstName, &a.LastName, &a.Date, &a.Time, &a.Group, &a.Subject, &a.Status); err != nil {
			return nil, err
		}
		results = append(results, a)
	}
	return results, nil
}
