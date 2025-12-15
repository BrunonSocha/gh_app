package main

import (
	"net/http"
	"github.com/justinas/alice"
	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	fileServer := http.FileServer(neuteredFileSystem{http.Dir("./ui/static/")})

	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	// sessions

	// handling functions
	router.HandlerFunc(http.MethodGet, "/", app.home)
	router.HandlerFunc(http.MethodGet, "/addinvoice", app.addInvoice)
	router.HandlerFunc(http.MethodPost, "/addinvoice", app.addInvoicePost)
	router.HandlerFunc(http.MethodGet, "/viewinvoice/:id", app.viewInvoice)
	router.HandlerFunc(http.MethodPost, "/jpk/create", app.jpkCreate)
	router.HandlerFunc(http.MethodGet, "/jpk/view/:id", app.jpkView)
	router.HandlerFunc(http.MethodPost, "/jpk/delete/:id", app.jpkDelete)
	router.HandlerFunc(http.MethodGet, "/jpk/viewall", app.jpkViewAll)
	router.HandlerFunc(http.MethodGet, "/viewinvoices", app.viewAllInvoices)
	router.HandlerFunc(http.MethodPost, "/deleteinvoice/:id", app.deleteInvoice)
	router.HandlerFunc(http.MethodGet, "/jpk/download/:id", app.jpkDownload)
	router.HandlerFunc(http.MethodPost, "/jpk/confirm/:id", app.jpkConfirm)

	standard := alice.New(app.sessionManager.LoadAndSave, app.recoverPanic, app.logRequest, secureHeaders)
	return standard.Then(router)
}

