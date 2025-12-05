package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"app.greyhouse.es/internal/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

	invoices, err := app.invoices.ThisMonth()
	if err != nil {
		app.serverError(w, err)
		return
	}

	files := []string{"ui/html/pages/home.tmpl", "ui/html/base.tmpl", "ui/html/partials/nav.tmpl"}


	templateset, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, err)
		return
	}
	tData := &templateData{
		Invoices: invoices,
	}

	err = templateset.ExecuteTemplate(w, "base", tData)
	if err != nil {
		app.serverError(w, err)
	}
}

func (app *application) addInvoice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	nr_faktury := "faktura 1488"
	nip := "1488148888"
	netto := 100
	podatek := 23
	data := time.Now()
	inv_type := models.PurchaseInvoice
	nazwa := "Nucysfera sp. z o.o."

	id, err := app.invoices.Insert(nip, nr_faktury, float64(netto), float64(podatek), data, inv_type, nazwa)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/viewinvoice?id=%d", id), http.StatusSeeOther)
}

func (app *application) viewInvoice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	inv, cname, err := app.invoices.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	
	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/pages/view.tmpl",
	}
	
	templateset, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, err)
		return
	}

	tData := &templateData{
		Invoice: inv,
		CompanyName: cname,
	}

	err = templateset.ExecuteTemplate(w, "base", tData)
	if err != nil {
		app.serverError(w, err)
	}

}

func (app *application) jpkView(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		app.notFound(w)
		return
	}
	fmt.Fprintf(w, "Wyświetl JPK o ID %d...", id)
}

func (app *application) jpkCreate(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Tworzenie pliku JPK..."))
}
