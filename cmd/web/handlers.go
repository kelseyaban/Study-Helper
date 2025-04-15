package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/abankelsey/study_helper/internal/data"
	"github.com/abankelsey/study_helper/internal/validator"
	"strconv"
)

// the home handles requests to display the home page
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// Initializes the template data for rendering the home page
	data := NewTemplateData()
	data.Title = "Home"
	data.HeaderText = "Welcome"

	// Render the home page template
	err := app.render(w, http.StatusOK, "home.tmpl", data)
	if err != nil {
		// Log the error and return Error response
		app.logger.Error("failed to render home page", "template", "home.tmpl", "error", err, "url", r.URL.Path, "method", r.Method)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// the showDailyGoals handles requests to display the daily goals form
func (app *application) showDailyGoalsForm(w http.ResponseWriter, r *http.Request) {
	// Initialize template data for the goals form
	data := NewTemplateData()
	data.Title = "Daily Goals"
	data.HeaderText = "Daily Goals"

	// Render the daily goals form template
	err := app.render(w, http.StatusOK, "daily_goals.tmpl", data)
	if err != nil {
		// Log the error and return Error response
		app.logger.Error("failed to render feedback page", "template", "daily_goals.tmpl", "error", err, "url", r.URL.Path, "method", r.Method)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (app *application) addGoals(w http.ResponseWriter, r *http.Request) {
	// Parse the submitted form data
	err := r.ParseForm()
	if err != nil {
		app.logger.Error("failed to parse form", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Extract form values
	goal_text := r.PostForm.Get("goal_text")
	is_completed_str := r.PostForm.Get("is_completed")
	target_date_str := r.PostForm.Get("target_date")

	// Convert the is_completed value from string to bool
	is_completed, err := strconv.ParseBool(is_completed_str)
	if err != nil {
		app.logger.Error("invalid value for is_completed", "value", is_completed_str)
		http.Error(w, "Invalid value for completion status", http.StatusBadRequest)
		return
	}

	// Convert target_date string to time.Time
	target_date, err := time.Parse("2006-01-02", target_date_str)
	if err != nil {
		app.logger.Error("invalid target_date format", "value", target_date_str)
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	// Create a goals object with the submitted data
	goals := &data.Goals{
		Goal_text:    goal_text,
		Is_completed: is_completed,
		Target_date:  target_date,
	}

	// Validate the submitted goals data
	v := validator.NewValidator()
	data.ValidateGoals(v, goals)

	// If validation fails, re-render the form with error messages
	if !v.ValidData() {
		data := NewTemplateData()
		data.Title = "Daily Goals"
		data.HeaderText = "Daily Goals"
		data.FormErrors = v.Errors         // Store validation errors
		data.FormData = map[string]string{ // Retain form input values
			"goal_text":    goal_text,
			"is_completed": is_completed_str,
			"target_date":  target_date_str,
		}

		// Render the form again with errors
		err := app.render(w, http.StatusUnprocessableEntity, "daily_goals.tmpl", data)
		if err != nil {
			app.logger.Error("failed to render daily goals page", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Insert the goal into the database
	err = app.goals.Insert(goals)
	if err != nil {
		app.logger.Error("failed to insert goal", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Redirect user to the success page
	http.Redirect(w, r, "/success", http.StatusSeeOther)
}

// the listGoals retrieves and displays all goal entries
func (app *application) listGoals(w http.ResponseWriter, r *http.Request) {
	// Fetch all gaol entries from the database
	goals, err := app.goals.GoalList() // fetches all stored goal sntries and return them as a list
	if err != nil {
		app.logger.Error("failed to fetch goals", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Prepare the template data with the retrieved journal entries
	data := NewTemplateData()
	data.Title = "Goal List"
	data.HeaderText = "All Goal Entries"
	data.GoalList = goals // Assign fetched goals entries to the template data

	// Render the goal list template
	err = app.render(w, http.StatusOK, "daily_goals_list.tmpl", data)
	if err != nil {
		app.logger.Error("failed to render goal list", "template", "daily_goals_list.tmpl", "error", err, "url", r.URL.Path, "method", r.Method)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// the deleteGoal will delete a goal from the database
func (app *application) deleteGoal(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	idStr := r.FormValue("goal_id")
	goalID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid goal ID", http.StatusBadRequest)
		return
	}

	err = app.goals.DeleteGoal(goalID)
	if err != nil {
		http.Error(w, "Could not delete goal", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/goals", http.StatusSeeOther)
}

// the showeditGoalForm handles requests to display the daily goals form to edit
func (app *application) showeditGoalForm(w http.ResponseWriter, r *http.Request) {
	// Get goal_id from query param
	goalIDStr := r.URL.Query().Get("goal_id")
	goalID, err := strconv.ParseInt(goalIDStr, 10, 64)
	if err != nil {
		app.logger.Error("invalid goal_id", "value", goalIDStr)
		http.Error(w, "Invalid goal ID", http.StatusBadRequest)
		return
	}

	// Fetch the goal from DB using goal_id
	goal, err := app.goals.GetGoalByID(goalID)
	if err != nil {
		app.logger.Error("failed to fetch goal for editing", "error", err)
		http.Error(w, "Could not find goal", http.StatusInternalServerError)
		return
	}

	// Preload the form with current goal values
	data := NewTemplateData()
	data.Title = "Edit Goal"
	data.HeaderText = "Edit Goal"
	data.FormData = map[string]string{
		"goal_id":      fmt.Sprintf("%d", goal.Goal_id),
		"goal_text":    goal.Goal_text,
		"is_completed": fmt.Sprintf("%t", goal.Is_completed),
		"target_date":  goal.Target_date.Format("2006-01-02"),
	}

	err = app.render(w, http.StatusOK, "edit_goal.tmpl", data)
	if err != nil {
		app.logger.Error("failed to render edit goal form", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// editGoal will update the  info from a goal entry
func (app *application) editGoal(w http.ResponseWriter, r *http.Request) {
	// Parse the submitted form data
	err := r.ParseForm()
	if err != nil {
		app.logger.Error("failed to parse form", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Extract the goal_id from the form
	goalIDStr := r.PostForm.Get("goal_id")
	goalID, err := strconv.ParseInt(goalIDStr, 10, 64)
	if err != nil {
		app.logger.Error("invalid goal_id", "value", goalIDStr)
		http.Error(w, "Invalid goal ID", http.StatusBadRequest)
		return
	}

	// Extract other form values
	goal_text := r.PostForm.Get("goal_text")
	is_completed_str := r.PostForm.Get("is_completed")
	target_date_str := r.PostForm.Get("target_date")

	// Convert the is_completed value from string to bool
	is_completed, err := strconv.ParseBool(is_completed_str)
	if err != nil {
		app.logger.Error("invalid value for is_completed", "value", is_completed_str)
		http.Error(w, "Invalid value for completion status", http.StatusBadRequest)
		return
	}

	// Convert target_date string to time.Time
	target_date, err := time.Parse("2006-01-02", target_date_str) // Standard date format (YYYY-MM-DD)
	if err != nil {
		app.logger.Error("invalid target_date format", "value", target_date_str)
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	// Create a goals object with the submitted data
	goals := &data.Goals{
		Goal_id:      goalID,
		Goal_text:    goal_text,
		Is_completed: is_completed,
		Target_date:  target_date,
	}

	// Validate the submitted goals data
	v := validator.NewValidator()
	data.ValidateGoals(v, goals)

	// If validation fails, re-render the form with error messages
	if !v.ValidData() {
		data := NewTemplateData()
		data.Title = "Edit Goal"
		data.HeaderText = "Edit Goal"
		data.FormErrors = v.Errors         // Store validation errors
		data.FormData = map[string]string{ // Retain form input values
			"goal_id":      goalIDStr,
			"goal_text":    goal_text,
			"is_completed": is_completed_str,
			"target_date":  target_date_str,
		}

		// Render the form again with errors
		err := app.render(w, http.StatusUnprocessableEntity, "edit_goal.tmpl", data)
		if err != nil {
			app.logger.Error("failed to render edit goal form", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Update the goal in the database
	err = app.goals.EditGoal(goals)
	if err != nil {
		app.logger.Error("failed to update goal", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Redirect user to the success page after updating
	http.Redirect(w, r, "/success", http.StatusSeeOther)
}

// the showSessionsForm handles requests to display the session form
func (app *application) showSessionsForm(w http.ResponseWriter, r *http.Request) {
	// Initialize template data for the session form
	data := NewTemplateData()
	data.Title = "Session"
	data.HeaderText = "Add a Session"

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

	// Construct sessions object
	sessions := &data.Sessions{
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
		data.Title = "Study Session"
		data.HeaderText = "Study Session"
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

	http.Redirect(w, r, "/success", http.StatusSeeOther)
}

// the listSessions retrieves and displays all session entries
func (app *application) listSessions(w http.ResponseWriter, r *http.Request) {
	// Fetch all gaol entries from the database
	sessions, err := app.sessions.SessionList() // fetches all stored session sntries and return them as a list
	if err != nil {
		app.logger.Error("failed to fetch session", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Prepare the template data with the retrieved journal entries
	data := NewTemplateData()
	data.Title = "Session List"
	data.HeaderText = "All Session Entries"
	data.SessionList = sessions // Assign fetched session entries to the template data

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

	err = app.sessions.DeleteSession(sessionID)
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

	http.Redirect(w, r, "/success", http.StatusSeeOther)
}

// the showQuoteForm handles requests to display the quote form
func (app *application) showQuoteForm(w http.ResponseWriter, r *http.Request) {
	// Initialize template data for the quote form
	data := NewTemplateData()
	data.Title = "Quote"
	data.HeaderText = "Add a Motivational Quote"

	// Render the quote form template
	err := app.render(w, http.StatusOK, "quotes.tmpl", data)
	if err != nil {
		// Log the error and return Error response
		app.logger.Error("failed to render quotes page", "template", "quotes.tmpl", "error", err, "url", r.URL.Path, "method", r.Method)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// the addQuote processes quote form submissions
func (app *application) addQuote(w http.ResponseWriter, r *http.Request) {
	// Parse the submitted form data
	err := r.ParseForm()
	if err != nil {
		app.logger.Error("failed to parse form", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Extract form values
	content := r.PostForm.Get("content")

	// Create a quote object with the submitted data
	quotes := &data.Quotes{
		Content: content,
	}

	// Validate the submitted quote data
	v := validator.NewValidator()
	data.ValidateQuotes(v, quotes)

	// If validation fails, re-render the form with error messages
	if !v.ValidData() {
		data := NewTemplateData()
		data.Title = "Quotes"
		data.HeaderText = "Submit Quote"
		data.FormErrors = v.Errors         // Store validation errors
		data.FormData = map[string]string{ // Retain form input values
			"content": content,
		}

		// Renders the quote form again with validation errors
		err := app.render(w, http.StatusUnprocessableEntity, "quotes.tmpl", data)
		if err != nil {
			app.logger.Error("failed to render quote page", "template", "quotes.tmpl", "error", err, "url", r.URL.Path, "method", r.Method)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Insert the quote into the database
	err = app.quotes.Insert(quotes)
	if err != nil {
		app.logger.Error("failed to insert quote", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Redirect user to the success page after successful submission
	http.Redirect(w, r, "/success", http.StatusSeeOther)
}

// the listQuotes handles requests to display a list of the submitted quote entries
func (app *application) listQuotes(w http.ResponseWriter, r *http.Request) {
	// Fetch all quote entries from the database
	quotes, err := app.quotes.QuoteList() // fetches all stored quote sntries and return them as a list
	if err != nil {
		app.logger.Error("failed to fetch quote", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Prepares the template data with the list of quote entries
	data := NewTemplateData()
	data.Title = "Quotes"
	data.HeaderText = "Quotes"
	data.QuoteList = quotes // Pass quote data to the template

	// Render the quote list template
	err = app.render(w, http.StatusOK, "quotes_list.tmpl", data)
	if err != nil {
		app.logger.Error("failed to render quote list", "template", "quotes_list.tmpl", "error", err, "url", r.URL.Path, "method", r.Method)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// the deleteQuote will delete a quote from the database
func (app *application) deleteQuote(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	idStr := r.FormValue("quote_id")
	quoteID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid quote ID", http.StatusBadRequest)
		return
	}

	err = app.quotes.DeleteQuote(quoteID)
	if err != nil {
		http.Error(w, "Could not delete quote", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/quotes", http.StatusSeeOther)
}

// Displays the success message page
func (app *application) showSuccessMessage(w http.ResponseWriter, r *http.Request) {
	// Create a new template data structure
	data := NewTemplateData()
	data.Title = "Submitted"
	data.HeaderText = "Submitted"

	// Render the success template
	err := app.render(w, http.StatusOK, "success.tmpl", data)
	if err != nil {
		// Log error if rendering fails
		app.logger.Error("failed to render success page", "template", "success.tmpl", "error", err, "url", r.URL.Path, "method", r.Method)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
