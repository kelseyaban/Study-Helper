package main

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
)

// bufferPool
var bufferPool = sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

// renders a template with the given data and sends the result to the client
func (app *application) render(w http.ResponseWriter, status int, page string, data *TemplateData) error {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf) // Return the buffer to the pool once finished

	// Check if the template exists in the cache
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("template %s does not exist", page)
		app.logger.Error("template does not exist", "template", page, "error", err)
		return err
	}

	// Execute the template with the provided data
	err := ts.Execute(buf, data)
	if err != nil {
		err = fmt.Errorf("failed to render template %s: %w", page, err)
		app.logger.Error("failed to render template", "template", page, "error", err)
		return err
	}

	// Write the response status and template content to the response
	w.WriteHeader(status)
	_, err = buf.WriteTo(w)
	if err != nil {
		err = fmt.Errorf("failed to write template to response: %w", err)
		app.logger.Error("failed to write template to response", "error", err)
		return err
	}

	return nil
}
