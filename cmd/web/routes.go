package main

import (
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))
	//The homepage
	mux.HandleFunc("GET /{$}", app.home)

	return app.loggingMiddleware(mux)
}
