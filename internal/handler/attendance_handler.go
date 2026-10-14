package handler

import (
	"fmt"
	"ku_backend_new/internal/domain"
	"ku_backend_new/internal/service"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/xuri/excelize/v2"
)

type AttendanceHandler struct {
	svc *service.AttendanceService
}

func NewAttendanceHandler(svc *service.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{svc: svc}
}

func (h *AttendanceHandler) ImportExcel(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "File not found"})
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	f, err := excelize.OpenReader(src)
	if err != nil {
		return err
	}

	rows, err := f.GetRows(f.GetSheetList()[0])
	if err != nil {
		return err
	}

	var records []domain.Absence
	for i, row := range rows {
		// Skip header and skip rows that don't have enough columns
		if i == 0 || len(row) < 6 {
			continue
		}

		// MAPPING TO YOUR EXCEL FILE:
		// A=0: Reg ID, B=1: Name, C=2: Date, D=3: Group, E=4: Subject, F=5: Status (yes/no)
		regID := strings.TrimSpace(row[0])
		fullName := strings.TrimSpace(row[1])
		dateStr := strings.TrimSpace(row[2])
		group := strings.TrimSpace(row[3])
		subject := strings.TrimSpace(row[4])
		rawStatus := strings.TrimSpace(row[5])

		// Field validation for critical fields
		if regID == "" || fullName == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("Row %d is missing required fields (Registration ID or Name)", i+1),
			})
		}

		// Logic to handle your "yes/no" status in the Excel
		status := "not_excused"
		if strings.ToLower(rawStatus) == "yes" {
			status = "excused"
		}

		// Parse the date (Matches your 2025-12-01 format)
		parsedDate, dateErr := time.Parse("2006-01-02", dateStr)
		if dateErr != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("Row %d has an invalid date format: '%s'. Expected YYYY-MM-DD", i+1, dateStr),
			})
		}

		// Split fullName into FirstName and LastName
		firstName := fullName
		lastName := ""
		nameParts := strings.SplitN(fullName, " ", 2)
		if len(nameParts) == 2 {
			firstName = nameParts[0]
			lastName = nameParts[1]
		}

		records = append(records, domain.Absence{
			RegNumber: regID,
			FirstName: firstName, 
			LastName:  lastName,
			Date:      parsedDate,
			Time:      "00:00",
			Group:     group,
			Subject:   subject,
			Status:    domain.AttendanceStatus(status),
		})
	}

	err = h.svc.ImportExcel(c.Request().Context(), records)
	if err != nil {
		log.Printf("Database Import Error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Successfully imported %d records", len(records)),
	})
}

func (h *AttendanceHandler) Fetch(c echo.Context) error {
    // 1. PULL the values from the URL
    monthStr := c.QueryParam("month")
    yearStr := c.QueryParam("year")
    group := c.QueryParam("group")
    search := c.QueryParam("search")

    // 2. CONVERT them to integers
    // We use Atoi to turn "12" into 12
    month, errM := strconv.Atoi(monthStr)
    year, errY := strconv.Atoi(yearStr)

    // 3. SAFETY CHECK: If conversion fails, default to current time
    if errM != nil || month == 0 {
        month = int(time.Now().Month())
    }
    if errY != nil || year == 0 {
        year = time.Now().Year()
    }

    // 4. LOG IT: This should now change when you click different months!
    fmt.Printf("HANDLER PARSED -> Month: %d, Year: %d from URL: %s\n", month, year, c.Request().RequestURI)

    // 5. SEND to Repository
    results, err := h.svc.GetFilteredAbsences(c.Request().Context(), group, search, month, year)
    if err != nil {
        return c.JSON(500, map[string]string{"error": err.Error()})
    }
    
    return c.JSON(200, results)
}

func (h *AttendanceHandler) Register(c echo.Context) error {
	var input domain.Absence

	// 1. Try to bind JSON from Flutter
	if err := c.Bind(&input); err != nil {
		fmt.Println("BIND ERROR:", err) // Watch your terminal for this!
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "JSON mapping failed"})
	}

	// 2. Log what arrived
	fmt.Printf("Received: %s for Subject: %s\n", input.FirstName, input.Subject)

	// 3. Try to save to DB
	err := h.svc.RegisterManual(c.Request().Context(), input)
	if err != nil {
		fmt.Println("DATABASE ERROR:", err) // Watch your terminal for this!
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "DB Save failed"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Success"})
}
