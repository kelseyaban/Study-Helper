package main

import (
	"errors"
	"math/rand"
	"net/http"
	"time"

	"github.com/abankelsey/study_helper/internal/data"
	"github.com/abankelsey/study_helper/internal/validator"
	"github.com/justinas/nosurf"
)

// the home handles requests to display the home page
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// Initializes the template data for rendering the home page
	data := NewTemplateData()
	data.Title = "Home"
	data.HeaderText = "Welcome"
	data.CSRFToken = nosurf.Token(r)

	userId := app.session.GetInt(r, "user_id")
	app.logger.Info("session user_id", "value", userId)

	// userId := app.session.GetInt(r.Context(), "user_id")
	if userId == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := int64(userId)
	app.logger.Info("user logged in", "user_id", userID)

	// Fetch all goals for display on the homepage
	goals, err := app.goals.GoalList(userID)
	if err != nil {
		app.logger.Error("failed to fetch goals for homepage", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	//  Assign to template data so they render on the home page
	data.GoalList = goals

	quotes, err := app.quotes.QuoteList(userID)
	if err != nil {
		app.logger.Error("failed to fetch quotes", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	//random quote if available
	if len(quotes) > 0 {
		rand.Seed(time.Now().UnixNano())       // seed random
		randomIndex := rand.Intn(len(quotes))  // pick random index
		data.RandomQuote = quotes[randomIndex] // assuming Quote is a struct
	}

	data.CurrentTime = time.Now()

	// Render the home page template
	err = app.render(w, http.StatusOK, "home.tmpl", data)
	if err != nil {
		app.logger.Error("failed to render home page", "template", "home.tmpl", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (app *application) isAuthenticated(r *http.Request) bool {
	return app.session.Exists(r, "authenticatedUserID")
}

func (app *application) showSignupForm(w http.ResponseWriter, r *http.Request) {
	// Initialize template data for the signup form
	data := NewTemplateData()
	data.Title = "Signup"
	data.HeaderText = "Signup"
	data.CSRFToken = nosurf.Token(r)

	// Render the daily goals form template
	err := app.render(w, http.StatusOK, "signup.tmpl", data)
	if err != nil {
		// Log the error and return Error response
		app.logger.Error("failed to render feedback page", "template", "signup.tmpl", "error", err, "url", r.URL.Path, "method", r.Method)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}

func (app *application) signupUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.logger.Error("failed to parse form", "error", err)
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.Form.Get("name")
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	// Create user instance
	users := &data.Users{
		Name:      name,
		Email:     email,
		Activated: true,
	}

	// Validate form data
	v := validator.NewValidator()
	data.ValidateUsers(v, users, password)

	// Show form again if validation failed
	if !v.ValidData() {
		data := NewTemplateData()
		data.Title = "Signup"
		data.HeaderText = "Study Helper"
		data.CSRFToken = nosurf.Token(r)
		data.FormErrors = v.Errors
		data.FormData = map[string]string{
			"name":  name,
			"email": email,
		}

		err := app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
		if err != nil {
			app.logger.Error("failed to render signup form", "template", "signup.tmpl", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Insert user into the database
	err = app.users.Insert(users, password)
	if err != nil {
		app.logger.Error("failed to insert user", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Redirect to login page or home page
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) showLoginForm(w http.ResponseWriter, r *http.Request) {
	// Initialize template data for the signup form
	data := NewTemplateData()
	data.Title = "Login"
	data.HeaderText = "Login"
	data.CSRFToken = nosurf.Token(r)

	// Render the daily goals form template
	err := app.render(w, http.StatusOK, "login.tmpl", data)
	if err != nil {
		// Log the error and return Error response
		app.logger.Error("failed to render feedback page", "template", "login.tmpl", "error", err, "url", r.URL.Path, "method", r.Method)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}

func (app *application) loginUser(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		app.logger.Error("failed to parse login form", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	// Validate input
	errors_user := make(map[string]string)
	if email == "" {
		errors_user["email"] = "Email is required"
	}
	if password == "" {
		errors_user["password"] = "Password is required"
	}

	// If validation errors, re-render form
	if len(errors_user) > 0 {
		data := NewTemplateData()
		data.Title = "Login"
		data.HeaderText = "Login"
		data.CSRFToken = nosurf.Token(r)
		data.FormErrors = errors_user
		data.FormData = map[string]string{
			"email": email,
		}

		err := app.render(w, http.StatusUnprocessableEntity, "login.tmpl", data)
		if err != nil {
			app.logger.Error("failed to render login form", "template", "login.tmpl", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Authenticate the user
	user, err := app.users.Authenticate(email, password)
	if err != nil {
		if errors.Is(err, data.ErrInvalidCredentials) {
			data := NewTemplateData()
			data.Title = "Login"
			data.HeaderText = "Login"
			data.CSRFToken = nosurf.Token(r)
			data.FormErrors = map[string]string{
				"generic": "Invalid email or password.",
			}
			data.FormData = map[string]string{
				"email": email,
			}

			err := app.render(w, http.StatusUnauthorized, "login.tmpl", data)
			if err != nil {
				app.logger.Error("failed to render login form", "template", "login.tmpl", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Unknown/internal error
		app.logger.Error("error authenticating user", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Store the user ID in the session
	app.session.Put(r, "user_id", int(user.User_id))
	app.session.Put(r, "authenticatedUserID", true)
	// app.logger.Info("user logged in", "user_id", user.User_id)

	// Redirect to the homepage
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) logoutUser(w http.ResponseWriter, r *http.Request) {
	app.session.Destroy(r)

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}
