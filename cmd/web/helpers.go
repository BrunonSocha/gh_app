package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
)

func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Println(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	templateSet, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("The template %s does not exist.", page)
		app.serverError(w, err)
		return
	}
	
	// prevent sending a 200 OK on template error
	
	buf := new(bytes.Buffer)
	err := templateSet.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}


	w.WriteHeader(status)

	buf.WriteTo(w)
}

func (app *application) newTemplateData(r *http.Request) *templateData {
	// for if we need to always pass in some data to the template
	return &templateData{}
}
