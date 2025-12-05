package main

import "net/http"

func (app *application) routes() *http.ServeMux {
		mux := http.NewServeMux()
	fileServer := http.FileServer(neuteredFileSystem{http.Dir("./ui/static/")})
	mux.Handle("/static", http.NotFoundHandler())
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))
	
	// handling functions
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/addinvoice", app.addInvoice)
	mux.HandleFunc("/viewinvoice", app.viewInvoice)
	mux.HandleFunc("/jpk/create", app.jpkCreate)
	mux.HandleFunc("/jpk/view", app.jpkView)

	return mux
}
