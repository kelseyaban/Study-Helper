package main

import (
	"log/slog"
	"net/http"
	"time"
)

func (app *application) serve() error {
	srv := &http.Server{
		Addr:         *app.addr,
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
		TLSConfig:    app.tlsConfig,
	}
	app.logger.Info("starting server", "addr", srv.Addr)
	return srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
}
