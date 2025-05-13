package main

import (
	"fmt"
	"net/http"
	"time"

	"strconv"

	"github.com/abankelsey/study_helper/internal/data"
	"github.com/abankelsey/study_helper/internal/validator"
	"github.com/justinas/nosurf"
)

// the showDailyGoals handles requests to display the daily goals form
func (app *application) showDailyGoalsForm(w http.ResponseWriter, r *http.Request) {
	// Initialize template data for the goals form
	data := NewTemplateData()
	data.Title = "Daily Goals"
	data.HeaderText = "Daily Goals"
	data.IsAuthenticated = app.isAuthenticated(r)
	data.CSRFToken = nosurf.Token(r)

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

	// Get the user ID from the session
	id := app.session.GetInt(r, "user_id")
	if id == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := int64(id)

	// Create a goals object with the submitted data
	goals := &data.Goals{
		Goal_text:    goal_text,
		Is_completed: is_completed,
		Target_date:  target_date,
		User_id:      userID,
	}

	// Validate the submitted goals data
	v := validator.NewValidator()
	data.ValidateGoals(v, goals)

	// If validation fails, re-render the form with error messages
	if !v.ValidData() {
		data := NewTemplateData()
		data.Title = "Daily Goals"
		data.HeaderText = "Daily Goals"
		data.IsAuthenticated = app.isAuthenticated(r)
		data.CSRFToken = nosurf.Token(r)
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

	//set session data
	app.session.Put(r, "flash", "Goal Successfully Added")

	// Redirect user to the goals page
	http.Redirect(w, r, "/goals", http.StatusSeeOther)
}

// the listGoals retrieves and displays all goal entries
func (app *application) listGoals(w http.ResponseWriter, r *http.Request) {

	// Get the user ID from the session
	id := app.session.GetInt(r, "user_id")
	if id == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := int64(id)

	// Fetch all gaol entries from the database
	goals, err := app.goals.GoalList(userID) // fetches all stored goal sntries and return them as a list
	if err != nil {
		app.logger.Error("failed to fetch goals", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	//Get/Check for the flash message
	flash := app.session.PopString(r, "flash")

	// Prepare the template data with the retrieved journal entries
	data := NewTemplateData()
	data.Title = "Goal List"
	data.HeaderText = "All Goal Entries"
	data.IsAuthenticated = app.isAuthenticated(r)
	data.CSRFToken = nosurf.Token(r)
	data.GoalList = goals // Assign fetched goals entries to the template data
	data.Flash = flash

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

	idStr := r.FormValue("goal_id")
	goalID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid goal ID", http.StatusBadRequest)
		return
	}

	err = app.goals.DeleteGoal(goalID, userID)
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
	data.IsAuthenticated = app.isAuthenticated(r)
	data.CSRFToken = nosurf.Token(r)
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
		data.IsAuthenticated = app.isAuthenticated(r)
		data.CSRFToken = nosurf.Token(r)
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

	// Redirect user to the goals page after updating
	http.Redirect(w, r, "/goals", http.StatusSeeOther)
}
