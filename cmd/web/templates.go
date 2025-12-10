package main

import (
	"html/template"
	"path/filepath"

	"app.greyhouse.es/internal/models"
)

type templateData struct {
	Invoice *models.Invoice
	Invoices []*models.Invoice
	CompanyName string
	Jpk *models.JPK
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		templateSet, err := template.ParseGlob("./ui/html/base.tmpl")
		if err != nil {
			return nil, err
		}

		templateSet, err = templateSet.ParseGlob("./ui/html/partials/*.tmpl")
		if err != nil {
			return nil, err
		}

		templateSet, err = templateSet.ParseFiles(page)
		if err != nil {
			return nil, err
		}
		cache[name] = templateSet
	}
	return cache, nil
} 

