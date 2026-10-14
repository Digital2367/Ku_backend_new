package repository

import (
	"context"
	"database/sql"
	"fmt"

	"ku_backend_new/internal/domain"
)

type SqliteAttendanceRepo struct {
	db *sql.DB
}

// NewSqliteAttendanceRepo creates a new instance of the repository and auto-creates the table
func NewSqliteAttendanceRepo(db *sql.DB) *SqliteAttendanceRepo {
	// Auto-create table so the user needs literally zero setup!
	createTableQuery := `
    CREATE TABLE IF NOT EXISTS absences (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        reg_number TEXT,
        first_name TEXT,
        last_name TEXT,
        absence_date DATETIME,
        absence_time TEXT,
        group_number TEXT,
        subject TEXT,
        status TEXT
    );`
	_, err := db.Exec(createTableQuery)
	if err != nil {
		fmt.Printf("Error creating SQLite table: %v\n", err)
	}

	return &SqliteAttendanceRepo{db: db}
}

// Save handles a single manual registration
func (r *SqliteAttendanceRepo) Save(ctx context.Context, record *domain.Absence) error {
	query := `INSERT INTO absences (reg_number, first_name, last_name, absence_date, absence_time, group_number, subject, status) 
              VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

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
func (r *SqliteAttendanceRepo) SaveBatch(ctx context.Context, records []domain.Absence) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO absences (reg_number, first_name, last_name, absence_date, absence_time, group_number, subject, status)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
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

func (r *SqliteAttendanceRepo) Find(ctx context.Context, group string, regNum string, month int, year int) ([]domain.Absence, error) {
	// SQLite date filtering using strftime
	query := `SELECT reg_number, first_name, last_name, absence_date, absence_time, group_number, subject, status
          FROM absences
          WHERE CAST(strftime('%m', absence_date) AS INTEGER) = ?
          AND CAST(strftime('%Y', absence_date) AS INTEGER) = ?`

	args := []interface{}{fmt.Sprintf("%02d", month), fmt.Sprintf("%d", year)}

	if group != "" && group != "All" {
		query += " AND group_number = ?"
		args = append(args, group)
	}

	if regNum != "" {
		query += " AND (reg_number LIKE ? OR first_name LIKE ? OR last_name LIKE ?)"
		likeStr := "%" + regNum + "%"
		args = append(args, likeStr, likeStr, likeStr)
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
