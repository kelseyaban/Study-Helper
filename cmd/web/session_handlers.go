package main

import (
	"fmt"
	"github.com/abankelsey/study_helper/internal/data"
	"github.com/abankelsey/study_helper/internal/validator"
	"github.com/justinas/nosurf"
	"net/http"
	"strconv"
	"time"
)

// the showSessionsForm handles requests to display the session form
func (app *application) showSessionsForm(w http.ResponseWriter, r *http.Request) {
	// Initialize template data for the session form
	data := NewTemplateData()
	data.Title = "Session"
	data.HeaderText = "Add a Session"
	data.IsAuthenticated = app.isAuthenticated(r)
	data.CSRFToken = nosurf.Token(r)

	// Render the sessions form template
	err := app.render(w, http.StatusOK, "sessions.tmpl", data)
	if err != nil {
		// Log the error and return Error response
		app.logger.Error("failed to render feedback page", "template", "sessions.tmpl", "error", err, "url", r.URL.Path, "method", r.Method)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (app *application) addSessions(w http.ResponseWriter, r *http.Request) {
	// Parse the submitted form data
	err := r.ParseForm()
	if err != nil {
		app.logger.Error("failed to parse form", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Extract form values
	title := r.Form.Get("title")
	description := r.Form.Get("description")
	subject := r.Form.Get("subject")
	start_date_str := r.Form.Get("start_date")
	end_date_str := r.Form.Get("end_date")
	isCompletedStr := r.Form.Get("is_completed")

	// Convert start_date string to time.Time
	start_date, err := time.Parse("2006-01-02", start_date_str)
	if err != nil {
		app.logger.Error("invalid start_date format", "value", start_date_str)
		http.Error(w, "Invalid start date format", http.StatusBadRequest)
		return
	}

	// Convert end_date string to time.Time
	end_date, err := time.Parse("2006-01-02", end_date_str)
	if err != nil {
		app.logger.Error("invalid end_date format", "value", end_date_str)
		http.Error(w, "Invalid end date format", http.StatusBadRequest)
		return
	}

	// Convert is_completed from string to bool
	isCompleted, err := strconv.ParseBool(isCompletedStr)
	if err != nil {
		app.logger.Error("invalid value for is_completed", "value", isCompletedStr)
		http.Error(w, "Invalid value for completion status", http.StatusBadRequest)
		return
	}
	// Get the user ID from the session
	id := app.session.GetInt(r, "user_id")
	if id == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := int64(id)

	// Construct sessions object
	sessions := &data.Sessions{
		Title:        title,
		Description:  description,
		Subject:      subject,
		Start_date:   start_date,
		End_date:     end_date,
		Is_completed: isCompleted,
		User_id:      userID,
	}

	// Validate
	v := validator.NewValidator()
	data.ValidateSessions(v, sessions)

	if !v.ValidData() {
		data := NewTemplateData()
		data.Title = "Study Session"
		data.HeaderText = "Study Session"
		data.IsAuthenticated = app.isAuthenticated(r)
		data.CSRFToken = nosurf.Token(r)
		data.FormErrors = v.Errors
		data.FormData = map[string]string{
			"title":        title,
			"description":  description,
			"subject":      subject,
			"start_date":   start_date_str,
			"end_date":     end_date_str,
			"is_completed": isCompletedStr,
		}

		err := app.render(w, http.StatusUnprocessableEntity, "sessions.tmpl", data)
		if err != nil {
			app.logger.Error("failed to render form", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Insert session
	err = app.sessions.Insert(sessions)
	if err != nil {
		app.logger.Error("failed to insert session", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	//set session data
	app.session.Put(r, "flash", "Session Successfully Added")

	http.Redirect(w, r, "/sessions", http.StatusSeeOther)
}

// the listSessions retrieves and displays all session entries
func (app *application) listSessions(w http.ResponseWriter, r *http.Request) {

	// Get userID from the session
	id := app.session.GetInt(r, "user_id")
	if id == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := int64(id)

	// Fetch all gaol entries from the database
	sessions, err := app.sessions.SessionList(userID) // fetches all stored session sntries and return them as a list
	if err != nil {
		app.logger.Error("failed to fetch session", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	//Get/Check for the flash message
	flash := app.session.PopString(r, "flash")

	// Prepare the template data with the retrieved journal entries
	data := NewTemplateData()
	data.Title = "Session List"
	data.HeaderText = "All Session Entries"
	data.IsAuthenticated = app.isAuthenticated(r)
	data.CSRFToken = nosurf.Token(r)
	data.SessionList = sessions // Assign fetched session entries to the template data
	data.Flash = flash

	// Render the session list template
	err = app.render(w, http.StatusOK, "sessions_list.tmpl", data)
	if err != nil {
		app.logger.Error("failed to render session list", "template", "sessions_list.tmpl", "error", err, "url", r.URL.Path, "method", r.Method)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// the deleteSession will delete a session from the database
func (app *application) deleteSession(w http.ResponseWriter, r *http.Request) {

	// Check and parse user ID from session
	id := app.session.GetInt(r, "user_id")
	if id == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := int64(id)

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	idStr := r.FormValue("session_id")
	sessionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	err = app.sessions.DeleteSession(sessionID, userID)
	if err != nil {
		http.Error(w, "Could not delete session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/sessions", http.StatusSeeOther)
}

// the showeditSessionForm handles requests to display the session form to edit
func (app *application) showeditSessionForm(w http.ResponseWriter, r *http.Request) {
	// Get session_id  from query param
	sessionIDStr := r.URL.Query().Get("session_id")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		app.logger.Error("invalid session_id", "value", sessionIDStr)
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Fetch the session from DB using session_id
	session, err := app.sessions.GetSessionByID(sessionID)
	if err != nil {
		app.logger.Error("failed to fetch session for editing", "error", err)
		http.Error(w, "Could not find session", http.StatusInternalServerError)
		return
	}

	// Preload the form with current session values
	data := NewTemplateData()
	data.Title = "Edit Session"
	data.HeaderText = "Edit Session"
	data.IsAuthenticated = app.isAuthenticated(r)
	data.CSRFToken = nosurf.Token(r)
	data.FormData = map[string]string{
		"session_id":   fmt.Sprintf("%d", session.Session_id),
		"title":        session.Title,
		"description":  session.Description,
		"subject":      session.Subject,
		"start_date":   session.Start_date.Format("2006-01-02"),
		"end_date":     session.End_date.Format("2006-01-02"),
		"is_completed": fmt.Sprintf("%t", session.Is_completed),
	}

	err = app.render(w, http.StatusOK, "edit_session.tmpl", data)
	if err != nil {
		app.logger.Error("failed to render edit session form", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (app *application) editSession(w http.ResponseWriter, r *http.Request) {
	// Parse the submitted form data
	err := r.ParseForm()
	if err != nil {
		app.logger.Error("failed to parse form", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Extract the session_id from the form
	sessionIDStr := r.PostForm.Get("session_id")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		app.logger.Error("invalid session_id", "value", sessionIDStr)
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Extract other form values
	title := r.PostForm.Get("title")
	description := r.PostForm.Get("description")
	subject := r.PostForm.Get("subject")
	start_date_str := r.PostForm.Get("start_date")
	end_date_str := r.PostForm.Get("end_date")
	is_completed_str := r.PostForm.Get("is_completed")

	// Convert start_date string to time.Time
	start_date, err := time.Parse("2006-01-02", start_date_str)
	if err != nil {
		app.logger.Error("invalid start_date format", "value", start_date_str)
		http.Error(w, "Invalid start date format", http.StatusBadRequest)
		return
	}

	// Convert end_date string to time.Time
	end_date, err := time.Parse("2006-01-02", end_date_str)
	if err != nil {
		app.logger.Error("invalid end_date format", "value", end_date_str)
		http.Error(w, "Invalid end date format", http.StatusBadRequest)
		return
	}

	// Convert is_completed from string to bool
	isCompleted, err := strconv.ParseBool(is_completed_str)
	if err != nil {
		app.logger.Error("invalid value for is_completed", "value", is_completed_str)
		http.Error(w, "Invalid value for completion status", http.StatusBadRequest)
		return
	}

	// Construct sessions object
	sessions := &data.Sessions{
		Session_id:   sessionID,
		Title:        title,
		Description:  description,
		Subject:      subject,
		Start_date:   start_date,
		End_date:     end_date,
		Is_completed: isCompleted,
	}

	// Validate
	v := validator.NewValidator()
	data.ValidateSessions(v, sessions)

	if !v.ValidData() {
		data := NewTemplateData()
		data.Title = "Edit Session"
		data.HeaderText = "Edit Session"
		data.IsAuthenticated = app.isAuthenticated(r)
		data.CSRFToken = nosurf.Token(r)
		data.FormErrors = v.Errors
		data.FormData = map[string]string{
			"session_id":   sessionIDStr,
			"title":        title,
			"description":  description,
			"subject":      subject,
			"start_date":   start_date_str,
			"end_date":     end_date_str,
			"is_completed": is_completed_str,
		}

		err := app.render(w, http.StatusUnprocessableEntity, "edit_session.tmpl", data)
		if err != nil {
			app.logger.Error("failed to render form", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Update  session
	err = app.sessions.EditSession(sessions)
	if err != nil {
		app.logger.Error("failed to insert session", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/sessions", http.StatusSeeOther)
}

// the showstartSessionInfo handles requests to display the session
func (app *application) showstartSessionInfo(w http.ResponseWriter, r *http.Request) {
	// Get session_id  from query param
	sessionIDStr := r.URL.Query().Get("session_id")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		app.logger.Error("invalid session_id", "value", sessionIDStr)
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Fetch the session from DB using session_id
	session, err := app.sessions.GetSessionByID(sessionID)
	if err != nil {
		app.logger.Error("failed to fetch session for editing", "error", err)
		http.Error(w, "Could not find session", http.StatusInternalServerError)
		return
	}

	// Preload the form with current session values
	data := NewTemplateData()
	data.Title = "Session Started"
	data.HeaderText = "Session Started"
	data.IsAuthenticated = app.isAuthenticated(r)
	data.CSRFToken = nosurf.Token(r)
	data.FormData = map[string]string{
		"session_id":   fmt.Sprintf("%d", session.Session_id),
		"title":        session.Title,
		"description":  session.Description,
		"subject":      session.Subject,
		"start_date":   session.Start_date.Format("2006-01-02"),
		"end_date":     session.End_date.Format("2006-01-02"),
		"is_completed": fmt.Sprintf("%t", session.Is_completed),
	}

	err = app.render(w, http.StatusOK, "session_start.tmpl", data)
	if err != nil {
		app.logger.Error("failed to render edit session form", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
