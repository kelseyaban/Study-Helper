package main

import (
	"net/http"
)

// the home handles requests to display the home page
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// Initializes the template data for rendering the home page
	data := NewTemplateData()
	data.Title = "Home"
	data.HeaderText = "Welcome to the Homepage"

	// Render the home page template
	err := app.render(w, http.StatusOK, "home.tmpl", data)
	if err != nil {
		// Log the error and return Error response
		app.logger.Error("failed to render home page", "template", "home.tmpl", "error", err, "url", r.URL.Path, "method", r.Method)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
