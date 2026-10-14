package main

import (
	"database/sql"
	"log"

	"ku_backend_new/internal/config"
	"ku_backend_new/internal/handler"
	"ku_backend_new/internal/repository"
	"ku_backend_new/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "modernc.org/sqlite" // Pure Go SQLite driver (No C compiler needed)
)

func main() {
	// 0. Load Configuration
	cfg := config.Load()

	// 1. Database Connection (Creates an attendance.db file automatically!)
	db, err := sql.Open("sqlite", "attendance.db")
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 2. Setup Layers (Dependency Injection)
	repo := repository.NewSqliteAttendanceRepo(db)
	svc := service.NewAttendanceService(repo)
	h := handler.NewAttendanceHandler(svc)

	e := echo.New()

	// 3. Middlewares
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}, latency=${latency_human}\n",
	}))
	e.Use(middleware.CORS())

	// 4. THE ROUTES (This fixes the 404!)
	e.GET("/attendance", h.Fetch)              // Matches uri=/attendance?month=1...
	e.POST("/attendance/register", h.Register) // Matches uri=/attendance/register
	// In main.go (where you set up routes)
	e.POST("/attendance/import", h.ImportExcel)

	// Mock Auth and Users Routes to fix 404 for Flutter App
	e.POST("/auth/login", func(c echo.Context) error {
		type LoginRequest struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		var req LoginRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(400, map[string]string{"error": "Invalid request"})
		}
		
		// For demo, accept anything and return Dean role
		return c.JSON(200, map[string]string{
			"role":     "Dean",
			"username": req.Username,
		})
	})

	e.GET("/users", func(c echo.Context) error {
		// Mock list of teachers
		users := []map[string]string{
			{"full_name": "John Doe", "role": "Teacher"},
			{"full_name": "Jane Smith", "role": "Dean"},
		}
		return c.JSON(200, users)
	})

	// Start Server
	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
