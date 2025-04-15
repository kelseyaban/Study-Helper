package main

import (
	"context"
	"database/sql"
	"flag"
	"html/template"
	"log/slog"
	"os"
	"time"

	// the '_' means that we will not direct use the pq package
	_ "github.com/lib/pq"

	"github.com/abankelsey/study_helper/internal/data"
)

// Dependency injection
type application struct {
	addr          *string
	goals         *data.GoalsModel
	logger        *slog.Logger // Logger for logging application events
	quotes        *data.QuotesModel
	sessions      *data.SessionsModel
	templateCache map[string]*template.Template // Cache for HTML templates
}

func main() {
	// Parsing command-line flags for the HTTP server address and DB connection string
	addr := flag.String("addr", "", "HTTP network address")
	dsn := flag.String("dsn", "", "PostgreSQL DSN")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Open the database connection
	db, err := openDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	logger.Info("database connection pool established")

	// Create a new template cache
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer db.Close()

	// Initialize the application with the dependencies
	app := &application{
		addr: addr,

		goals:         &data.GoalsModel{DB: db},
		logger:        logger,
		quotes:        &data.QuotesModel{DB: db},
		sessions:      &data.SessionsModel{DB: db},
		templateCache: templateCache,
	}

	// Start the application server
	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

// theh openDB opens a connection to the database and verifies the connection
func openDB(dsn string) (*sql.DB, error) {
	// Opens the PostgreSQL database
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// A context with 5-second timeout for the database ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check the database connection
	err = db.PingContext(ctx)
	if err != nil {
		db.Close() // Close the connection on error
		return nil, err
	}

	return db, nil // Return the established connection
}
