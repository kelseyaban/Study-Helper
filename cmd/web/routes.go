package main

import (
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))
	// //The login page
	// mux.HandleFunc("GET /", app.login)
	// //Handle user login
	// mux.HandleFunc("POST /", app.addUser)

	//the home page
	mux.HandleFunc("GET /", app.home)

	//Handle daily goals form
	mux.HandleFunc("GET /goal", app.showDailyGoalsForm)
	//Handle daily goals submissions
	mux.HandleFunc("POST /goal", app.addGoals)
	//Get all goal entries
	mux.HandleFunc("GET /goals", app.listGoals)

	//Handle daily goals form
	mux.HandleFunc("GET /session", app.showSessionsForm)

	mux.HandleFunc("GET /success", app.showSuccessMessage)

	return app.loggingMiddleware(mux)
}
