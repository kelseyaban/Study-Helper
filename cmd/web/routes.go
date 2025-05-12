package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	dynamicMiddleware := alice.New(app.session.LoadAndSave)

	//signup
	mux.Handle("GET /user/signup", dynamicMiddleware.ThenFunc(app.showSignupForm))
	mux.Handle("POST /user/signup", dynamicMiddleware.ThenFunc(app.signupUser))

	mux.Handle("GET /user/login", dynamicMiddleware.ThenFunc(app.showLoginForm))
	mux.Handle("POST /user/login", dynamicMiddleware.ThenFunc(app.loginUser))
	mux.Handle("POST /user/logout", dynamicMiddleware.ThenFunc(app.logoutUser))

	//the home page
	mux.Handle("GET /", dynamicMiddleware.ThenFunc(app.home))

	//Handle daily goals form
	mux.Handle("GET /goal", dynamicMiddleware.ThenFunc(app.showDailyGoalsForm))
	//Handle daily goals submissions
	mux.Handle("POST /goal", dynamicMiddleware.ThenFunc(app.addGoals))
	//Get all goal entries
	mux.Handle("GET /goals", dynamicMiddleware.ThenFunc(app.listGoals))
	//Handle delete a goal
	mux.Handle("POST /goals/delete", dynamicMiddleware.ThenFunc(app.deleteGoal))
	//Handle edit goal form
	mux.Handle("GET /goals/edit", dynamicMiddleware.ThenFunc(app.showeditGoalForm))
	//Hnalde the edit goal
	mux.Handle("POST /goals/edit", dynamicMiddleware.ThenFunc(app.editGoal))

	//Handle study sessions form
	mux.Handle("GET /session", dynamicMiddleware.ThenFunc(app.showSessionsForm))
	//Handle study session submissions
	mux.Handle("POST /session", dynamicMiddleware.ThenFunc(app.addSessions))
	//Get all session entries
	mux.Handle("GET /sessions", dynamicMiddleware.ThenFunc(app.listSessions))
	//Handle delete a session
	mux.Handle("POST /sessions/delete", dynamicMiddleware.ThenFunc(app.deleteSession))
	//Handle edit session form
	mux.Handle("GET /sessions/edit", dynamicMiddleware.ThenFunc(app.showeditSessionForm))
	//Hnalde the edit session
	mux.Handle("POST /sessions/edit", dynamicMiddleware.ThenFunc(app.editSession))
	//Handle show session form
	mux.Handle("GET /sessions/start", dynamicMiddleware.ThenFunc(app.showstartSessionInfo))

	//Handle quote form
	mux.Handle("GET /quote", dynamicMiddleware.ThenFunc(app.showQuoteForm))
	//Handle quote submissions
	mux.Handle("POST /quote", dynamicMiddleware.ThenFunc(app.addQuote))
	//Get all quote entries
	mux.Handle("GET /quotes", dynamicMiddleware.ThenFunc(app.listQuotes))
	//Handle delete a quote
	mux.Handle("POST /quotes/delete", dynamicMiddleware.ThenFunc(app.deleteQuote))

	return app.loggingMiddleware(mux)
}
