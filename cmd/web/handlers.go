package main

import (
	"encoding/xml"
	"errors"
	"fmt"
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

	invoices, err := app.invoices.LastMonth()
	if err != nil {
		app.serverError(w, err)
		return
	}
	
	app.render(w, http.StatusOK, "home.tmpl", &templateData{Invoices: invoices})
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
	data := time.Now().AddDate(0, -1, 0)
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
	
	app.render(w, http.StatusOK, "view_invoice.tmpl", &templateData{Invoice: inv, CompanyName: cname})
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
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}
	invoices, err := app.invoices.LastMonth()
	if err != nil {
		app.serverError(w, err)
		return
	}

	jpk, err := app.jpks.NewJpk(invoices)
	if err != nil {
		app.serverError(w, err)
		return
	}
	Header := `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
	out, err := xml.MarshalIndent(jpk, "", "  ")
	if err != nil {
		app.serverError(w, err)
		return
	}
	out = []byte(Header + string(out))

	fmt.Fprintf(w, string(out))

}
