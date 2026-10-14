package service

import (
	"context"
	"fmt"
	// "io"
	"time"

	"ku_backend_new/internal/domain"
	// "github.com/xuri/excelize/v2"
)

type AttendanceService struct {
	repo domain.AttendanceRepository
}

func NewAttendanceService(r domain.AttendanceRepository) *AttendanceService {
	return &AttendanceService{repo: r}
}

// 1. Manual Registration (New Registering)
func (s *AttendanceService) RegisterManual(ctx context.Context, abs domain.Absence) error {
	if abs.RegNumber == "" || abs.FirstName == "" {
		return fmt.Errorf("registration number and name are required")
	}
	// Default status if empty
	if abs.Status == "" {
		abs.Status = domain.StatusNotExcused
	}
	// Default date to today if missing
	if abs.Date.IsZero() {
		abs.Date = time.Now()
	}
	return s.repo.Save(ctx, &abs)
}

// 2. Search & Month Selection
func (s *AttendanceService) GetAttendance(ctx context.Context, group, reg string, month, year int) ([]domain.Absence, error) {
	// If the user doesn't provide a year, assume the current one
	if year == 0 {
		year = time.Now().Year()
	}
	return s.repo.Find(ctx, group, reg, month, year)
}
func (s *AttendanceService) GetFilteredAbsences(ctx context.Context, group, search string, month, year int) ([]domain.Absence, error) {
	if year == 0 {
		year = time.Now().Year()
	}
	if month == 0 {
		month = int(time.Now().Month())
	}
	return s.repo.Find(ctx, group, search, month, year)
}

// 3. Excel Import

func (s *AttendanceService) ImportExcel(ctx context.Context, records []domain.Absence) error {
	// This calls the SaveBatch method we added to your Repository earlier
	return s.repo.SaveBatch(ctx, records)
}
