package main

import (
	"html/template"
	"path/filepath"
)

// loads all template files and returns a map with the template file name as the key and the parsed template as the value.
func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	// Get all template files in the ui/html directory
	pages, err := filepath.Glob("./ui/html/*.tmpl")
	if err != nil {
		return nil, err
	}

	// Parse each template file and store it in the cache
	for _, page := range pages {
		fileName := filepath.Base(page)
		ts, err := template.ParseFiles(page)
		if err != nil {
			return nil, err
		}
		cache[fileName] = ts
	}

	return cache, nil
}
