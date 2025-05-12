package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	// the '_' means that we will not direct use the pq package
	"github.com/abankelsey/study_helper/internal/data"
	"github.com/alexedwards/scs/v2"
	_ "github.com/lib/pq"
)

// Dependency injection
type application struct {
	addr          *string
	goals         *data.GoalsModel
	logger        *slog.Logger // Logger for logging application events
	quotes        *data.QuotesModel
	sessions      *data.SessionsModel
	session       *scs.SessionManager
	templateCache map[string]*template.Template // Cache for HTML templates
	tlsConfig     *tls.Config
	users         *data.UsersModel
}

func main() {
	// Parsing command-line flags for the HTTP server address and DB connection string
	addr := flag.String("addr", "", "HTTP network address")
	dsn := flag.String("dsn", "", "PostgreSQL DSN")
	// secret := flag.String("secret", "KidajE20eufaLsfdS*20+jEhrwrw_uYh", "Secret key")

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

	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	sessionManager.Cookie.Secure = false // ⚠️ only for local dev

	//ECDHE -
	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	// Initialize the application with the dependencies
	app := &application{
		addr:          addr,
		goals:         &data.GoalsModel{DB: db},
		logger:        logger,
		quotes:        &data.QuotesModel{DB: db},
		sessions:      &data.SessionsModel{DB: db},
		templateCache: templateCache,
		session:       sessionManager,
		tlsConfig:     tlsConfig,
		users:         &data.UsersModel{DB: db},
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
