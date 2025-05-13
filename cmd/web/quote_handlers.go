package main

import (
	"github.com/abankelsey/study_helper/internal/data"
	"github.com/abankelsey/study_helper/internal/validator"
	"github.com/justinas/nosurf"
	"net/http"
	"strconv"
)

// the showQuoteForm handles requests to display the quote form
func (app *application) showQuoteForm(w http.ResponseWriter, r *http.Request) {
	// Initialize template data for the quote form
	data := NewTemplateData()
	data.Title = "Quote"
	data.HeaderText = "Add a Motivational Quote"
	data.IsAuthenticated = app.isAuthenticated(r)
	data.CSRFToken = nosurf.Token(r)

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

	// Get the user ID from the session
	id := app.session.GetInt(r, "user_id")
	if id == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := int64(id)

	// Create a quote object with the submitted data and user ID
	quotes := &data.Quotes{
		Content: content,
		User_id: userID,
	}

	// Validate the submitted quote data
	v := validator.NewValidator()
	data.ValidateQuotes(v, quotes)

	// If validation fails, re-render the form with error messages
	if !v.ValidData() {
		data := NewTemplateData()
		data.Title = "Quotes"
		data.HeaderText = "Quotes"
		data.IsAuthenticated = app.isAuthenticated(r)
		data.CSRFToken = nosurf.Token(r)
		data.FormErrors = v.Errors
		data.FormData = map[string]string{
			"content": content,
		}

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

	// Set session flash message
	app.session.Put(r, "flash", "Quote successfully added")

	// Redirect user to the quotes page
	http.Redirect(w, r, "/quotes", http.StatusSeeOther)
}

// the listQuotes handles requests to display a list of the submitted quote entries
func (app *application) listQuotes(w http.ResponseWriter, r *http.Request) {
	// Get userID from the session
	id := app.session.GetInt(r, "user_id")
	if id == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := int64(id)

	// Fetch quotes for the current user
	quotes, err := app.quotes.QuoteList(userID)
	if err != nil {
		app.logger.Error("failed to fetch quotes", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	//Get/Check for the flash message
	flash := app.session.PopString(r, "flash")

	// Prepares the template data with the list of quote entries
	data := NewTemplateData()
	data.Title = "Quotes"
	data.HeaderText = "Quotes"
	data.IsAuthenticated = app.isAuthenticated(r)
	data.CSRFToken = nosurf.Token(r)
	data.QuoteList = quotes // Pass quote data to the template
	data.Flash = flash

	// Render the quote list template
	err = app.render(w, http.StatusOK, "quotes_list.tmpl", data)
	if err != nil {
		app.logger.Error("failed to render quote list", "template", "quotes_list.tmpl", "error", err, "url", r.URL.Path, "method", r.Method)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (app *application) deleteQuote(w http.ResponseWriter, r *http.Request) {
	// Check and parse user ID from session
	id := app.session.GetInt(r, "user_id")
	if id == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := int64(id)

	// Parse form to get quote ID
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

	// Call DeleteQuote with both quoteID and userID
	err = app.quotes.DeleteQuote(quoteID, userID)
	if err != nil {
		http.Error(w, "Could not delete quote", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/quotes", http.StatusSeeOther)
}
