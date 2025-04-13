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
	// Initialize template data for the feedback form
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

// the showDailyGoals handles requests to display the daily goals form
func (app *application) showSessionsForm(w http.ResponseWriter, r *http.Request) {
	// Initialize template data for the feedback form
	data := NewTemplateData()
	data.Title = "Session"
	data.HeaderText = "Add a Session"

	// Render the daily goals form template
	err := app.render(w, http.StatusOK, "sessions.tmpl", data)
	if err != nil {
		// Log the error and return Error response
		app.logger.Error("failed to render feedback page", "template", "sessions.tmpl", "error", err, "url", r.URL.Path, "method", r.Method)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
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

//You can ignore this functions, I was just messing around with the login
// func (app *application) login(w http.ResponseWriter, r *http.Request) {
// 	// Initializes the template data for rendering the home page
// 	data := NewTemplateData()
// 	data.Title = "Login"
// 	data.HeaderText = "Login"

// 	// Render the home page template
// 	err := app.render(w, http.StatusOK, "login.tmpl", data)
// 	if err != nil {
// 		// Log the error and return Error response
// 		app.logger.Error("failed to render home page", "template", "login.tmpl", "error", err, "url", r.URL.Path, "method", r.Method)
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 		return
// 	}
// }

// // the addUser processes login form submissions
// func (app *application) addUser(w http.ResponseWriter, r *http.Request) {
// 	// // Parse the submitted form data
// 	// err := r.ParseForm()
// 	// if err != nil {
// 	// 	app.logger.Error("failed to parse form", "error", err)
// 	// 	http.Error(w, "Bad Request", http.StatusBadRequest)
// 	// 	return
// 	// }

// 	// // Extract form values
// 	// username := r.PostForm.Get("username")
// 	// email := r.PostForm.Get("email")
// 	// password_hash := r.PostForm.Get("password_hash")

// 	// // Create a feedback object with the submitted data
// 	// users := &data.Users{
// 	// 	Username:      username,
// 	// 	Email:         email,
// 	// 	Password_hash: password_hash,
// 	// }

// 	// // Validate the submitted feedback data
// 	// v := validator.NewValidator()
// 	// data.ValidateUsers(v, users)

// 	// // If validation fails, re-render the form with error messages
// 	// if !v.ValidData() {
// 	// 	data := NewTemplateData()
// 	// 	data.Title = "Login"
// 	// 	data.HeaderText = "Login"
// 	// 	data.FormErrors = v.Errors         // Store validation errors
// 	// 	data.FormData = map[string]string{ // Retain form input values
// 	// 		"username":      username,
// 	// 		"email":         email,
// 	// 		"password_hash": password_hash,
// 	// 	}

// 	// 	// Renders the feedback form again with validation errors
// 	// 	err := app.render(w, http.StatusUnprocessableEntity, "login.tmpl", data)
// 	// 	if err != nil {
// 	// 		app.logger.Error("failed to render feedback page", "template", "login.tmpl", "error", err, "url", r.URL.Path, "method", r.Method)
// 	// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 	// 		return
// 	// 	}
// 	// 	return
// 	// }

// 	// // Insert the feedback into the database
// 	// err = app.users.Insert(users)
// 	// if err != nil {
// 	// 	app.logger.Error("failed to insert feedback", "error", err)
// 	// 	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 	// 	return
// 	// }

// 	// Redirect user to the success page after successful submission
// 	http.Redirect(w, r, "/home", http.StatusSeeOther)
// }
