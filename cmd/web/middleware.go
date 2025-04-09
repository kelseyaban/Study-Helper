package main

import (
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
