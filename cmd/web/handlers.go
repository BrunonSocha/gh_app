package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"app.greyhouse.es/internal/models"
	"github.com/julienschmidt/httprouter"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {

	invoices, err := app.invoices.LastMonth()
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Invoices = invoices
	
	app.render(w, http.StatusOK, "home.tmpl", data)
}

func (app *application) addInvoice(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Show the form to create a new invoice.."))
}

func (app *application) addInvoicePost(w http.ResponseWriter, r *http.Request) {

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

	http.Redirect(w, r, fmt.Sprintf("/viewinvoice/%d", id), http.StatusSeeOther)
}

func (app *application) viewInvoice(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	inv, cname, err := app.invoices.Get(id)
	data := app.newTemplateData(r)
	data.Invoice = inv
	data.CompanyName = cname
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	
	app.render(w, http.StatusOK, "view_invoice.tmpl", data)
}

func (app *application) jpkView(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.notFound(w)
		return
	}
	data := app.newTemplateData(r)
	data.Jpk, data.JpkMetadata, err = app.jpks.Get(id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.render(w, http.StatusOK, "view_jpk.tmpl", data)
}

func (app *application) jpkCreate(w http.ResponseWriter, r *http.Request) {
	invoices, err := app.invoices.LastMonth()
	if err != nil {
		app.serverError(w, err)
		return
	}

	jpk, err := app.jpks.NewJpk(invoices)

	// add the jpk to a newTemplateData. Create a template for displaying jpks.
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
	// redirect to view jpk.
	fmt.Fprintf(w, string(out))

}

func (app *application) jpkDelete(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	err = app.jpks.Delete(id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	fmt.Fprintf(w, "Usunięto JPK o id %d", params)
}
