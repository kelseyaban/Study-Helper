package main

import (
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	//the home page
	mux.HandleFunc("GET /", app.home)

	//Handle daily goals form
	mux.HandleFunc("GET /goal", app.showDailyGoalsForm)
	//Handle daily goals submissions
	mux.HandleFunc("POST /goal", app.addGoals)
	//Get all goal entries
	mux.HandleFunc("GET /goals", app.listGoals)
	//Handle delete a goal
	mux.HandleFunc("POST /goals/delete", app.deleteGoal)
	//Handle edit goal form
	mux.HandleFunc("GET /goals/edit", app.showeditGoalForm)
	//Hnalde the edit goal
	mux.HandleFunc("POST /goals/edit", app.editGoal)

	//Handle study sessions form
	mux.HandleFunc("GET /session", app.showSessionsForm)
	//Handle study session submissions
	mux.HandleFunc("POST /session", app.addSessions)
	//Get all session entries
	mux.HandleFunc("GET /sessions", app.listSessions)
	//Handle delete a session
	mux.HandleFunc("POST /sessions/delete", app.deleteSession)
	//Handle edit session form
	mux.HandleFunc("GET /sessions/edit", app.showeditSessionForm)
	//Hnalde the edit session
	mux.HandleFunc("POST /sessions/edit", app.editSession)

	mux.HandleFunc("GET /success", app.showSuccessMessage)

	return app.loggingMiddleware(mux)
}
