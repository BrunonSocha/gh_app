package main

import (
	"net/http"
	"github.com/justinas/alice"
	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	
	router := httprouter.New()

	fileServer := http.FileServer(neuteredFileSystem{http.Dir("./ui/static/")})

	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	// handling functions
	router.HandlerFunc(http.MethodGet, "/", app.home)
	router.HandlerFunc(http.MethodGet, "/addinvoice", app.addInvoice)
	router.HandlerFunc(http.MethodPost, "addinvoice", app.addInvoicePost)
	router.HandlerFunc(http.MethodGet, "/viewinvoice/:id", app.viewInvoice)
	router.HandlerFunc(http.MethodPost, "/jpk/create", app.jpkCreate)
	router.HandlerFunc(http.MethodGet, "/jpk/view", app.jpkView)
	
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	return standard.Then(mux)
}

