package domain

import (
	"context"
	"time"
)

type AttendanceStatus string

const (
	StatusExcused    AttendanceStatus = "excused"
	StatusNotExcused AttendanceStatus = "not_excused"
)

// internal/domain/attendance.go
type Absence struct {
	RegNumber string           `json:"reg_number"`
	FirstName string           `json:"first_name"`
	LastName  string           `json:"last_name"`
	Date      time.Time        `json:"date"`
	Time      string           `json:"absence_time"`
	Group     string           `json:"group_number"`
	Subject   string           `json:"subject"`
	Status    AttendanceStatus `json:"status"`
}

type AttendanceRepository interface {
	Save(ctx context.Context, record *Absence) error
	SaveBatch(ctx context.Context, records []Absence) error
	// The Find method for searching and month filtering
	Find(ctx context.Context, group string, regNum string, month int, year int) ([]Absence, error)
}
