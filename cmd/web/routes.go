package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"net/http"
)

func (app *application) routes() http.Handler {

	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	fileServer := http.FileServer(neuteredFileSystem{http.Dir("./ui/static/")})

	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	dynamic := alice.New(app.sessionManager.LoadAndSave)

	// sessions

	// handling functions
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/addinvoice", dynamic.ThenFunc(app.addInvoice))
	router.Handler(http.MethodPost, "/addinvoice", dynamic.ThenFunc(app.addInvoicePost))
	router.Handler(http.MethodGet, "/viewinvoice/:id", dynamic.ThenFunc(app.viewInvoice))
	router.Handler(http.MethodPost, "/jpk/create", dynamic.ThenFunc(app.addJpk))
	router.Handler(http.MethodGet, "/jpk/view/:id", dynamic.ThenFunc(app.viewJpk))
	router.Handler(http.MethodPost, "/jpk/delete/:id", dynamic.ThenFunc(app.deleteJpk))
	router.Handler(http.MethodGet, "/jpk/viewall", dynamic.ThenFunc(app.viewAllJpk))
	router.Handler(http.MethodGet, "/viewinvoices", dynamic.ThenFunc(app.viewAllInvoices))
	router.Handler(http.MethodPost, "/deleteinvoice/:id", dynamic.ThenFunc(app.deleteInvoice))
	router.Handler(http.MethodGet, "/jpk/download/:id", dynamic.ThenFunc(app.downloadJpk))
	router.Handler(http.MethodPost, "/jpk/confirm/:id", dynamic.ThenFunc(app.confirmJpk))
	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.userSignUp))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.userSignUpPost))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.userLogin))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.userLoginPost))
	router.Handler(http.MethodPost, "/user/logout", dynamic.ThenFunc(app.userLogoutPost))
	standard := alice.New(app.sessionManager.LoadAndSave, app.recoverPanic, app.logRequest, secureHeaders)
	return standard.Then(router)
}
