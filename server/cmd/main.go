package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samokw/ssl_tracker/internal/database"
	"github.com/samokw/ssl_tracker/internal/domain"
	"github.com/samokw/ssl_tracker/internal/ssl"
	"github.com/samokw/ssl_tracker/internal/tui"
)

// Creating a basic program that will check the exipry of a predefined sercer
func main() {
	// Disable logging for TUI mode to prevent console output interference
	logger := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{
		Level:     slog.LevelError, // Only log errors, and discard them
		AddSource: false,
	}))
	slog.SetDefault(logger)

	// Initialize database
	dbPath, err := database.GetDefaultDBPath()
	if err != nil {
		fmt.Printf("Error getting database path: %v\n", err)
		os.Exit(1)
	}

	db, err := database.InitSQLite(dbPath)
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	domainRepo := domain.NewRepository(db)
	sslService := ssl.NewCertService()
	domainService := domain.NewService(domainRepo, sslService)

	app := tui.NewApp(domainService)
	program := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
