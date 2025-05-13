package main

import (
	"github.com/justinas/nosurf"
	"net/http"
)

// logs incoming HTTP requests and response details
func (app *application) loggingMiddleware(next http.Handler) http.Handler {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Captures the  details of the request
		ip := r.RemoteAddr
		proto := r.Proto
		method := r.Method
		uri := r.URL.RequestURI()

		// Log the incoming request
		app.logger.Info("received request", "ip", ip, "protocol", proto, "method", method, "uri", uri)

		// Calls the next handler
		next.ServeHTTP(w, r)

		// Log after the request is processed
		app.logger.Info("Request processed")
	})
	return fn
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Use the isAuthenticated
		if !app.isAuthenticated(r) {
			app.logger.Warn("Authentication required", "uri", r.URL.RequestURI()) // Log attempt

			// Redirect the user to the login page.
			http.Redirect(w, r, "/user/login", http.StatusFound)
			return
		}

		w.Header().Add("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// noSurf middleware adds CSRF protection to all non-safe methods like POST, PUT, DELETE
func noSurf(next http.Handler) http.Handler {
	// Create a new CSRF handler
	csrfHandler := nosurf.New(next)

	// Configure the base cookie settings
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",                  // Available across the entire site
		Secure:   true,                 // Requires HTTPS
		SameSite: http.SameSiteLaxMode, // Standard SameSite setting
	})

	return csrfHandler
}
