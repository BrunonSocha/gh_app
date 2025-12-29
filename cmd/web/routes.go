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

	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)
	protected := dynamic.Append(app.requireAuthentication, app.requireNIP)
	// sessions

	// protected
	router.Handler(http.MethodGet, "/", protected.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/addinvoice", protected.ThenFunc(app.addInvoice))
	router.Handler(http.MethodPost, "/addinvoice", protected.ThenFunc(app.addInvoicePost))
	router.Handler(http.MethodGet, "/viewinvoice/:id", protected.ThenFunc(app.viewInvoice))
	router.Handler(http.MethodPost, "/jpk/create", protected.ThenFunc(app.addJpk))
	router.Handler(http.MethodGet, "/jpk/view/:id", protected.ThenFunc(app.viewJpk))
	router.Handler(http.MethodPost, "/jpk/delete/:id", protected.ThenFunc(app.deleteJpk))
	router.Handler(http.MethodGet, "/jpk/viewall", protected.ThenFunc(app.viewAllJpk))
	router.Handler(http.MethodGet, "/viewinvoices", protected.ThenFunc(app.viewAllInvoices))
	router.Handler(http.MethodPost, "/deleteinvoice/:id", protected.ThenFunc(app.deleteInvoice))
	router.Handler(http.MethodGet, "/jpk/download/:id", protected.ThenFunc(app.downloadJpk))
	router.Handler(http.MethodPost, "/jpk/confirm/:id", protected.ThenFunc(app.confirmJpk))
	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.userLogoutPost))
	//

	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.userSignUp))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.userSignUpPost))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.userLogin))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.userLoginPost))
	standard := alice.New(app.sessionManager.LoadAndSave, app.recoverPanic, app.logRequest, secureHeaders)
	return standard.Then(router)
}
