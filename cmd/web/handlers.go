package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"app.greyhouse.es/internal/models"
	"github.com/julienschmidt/httprouter"
)

type addInvoiceForm struct {
	Nr_faktury string
	NIP string
	Nazwa string
	Netto float64
	Podatek float64
	Data time.Time
	Inv_type models.InvoiceType
	FieldErrors map[string]string
}

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
	data := app.newTemplateData(r)
	data.Form = addInvoiceForm{}
	app.render(w, http.StatusOK, "add_invoice.tmpl", data)
}

func (app *application) addInvoicePost(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	netto, err := strconv.ParseFloat(r.PostForm.Get("netto"), 64)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	podatek, err := strconv.ParseFloat(r.PostForm.Get("podatek"), 64)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	data, err := time.Parse("2006-01-02",r.PostForm.Get("data"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	var inv_type models.InvoiceType
	if r.PostForm.Get("type") == "PURC" {
		inv_type = models.PurchaseInvoice
	} else {
		inv_type = models.SaleInvoice
	}
	nazwa := r.PostForm.Get("nazwa")
	form := addInvoiceForm {
		Nr_faktury: r.PostForm.Get("nr_faktury"),
		NIP: strings.ReplaceAll(strings.ReplaceAll(r.PostForm.Get("nip"), "-", ""), " ", ""),
		Netto: netto,
		Podatek: podatek,
		Data: data,
		Nazwa: nazwa,
		Inv_type: inv_type,
		FieldErrors: make(map[string]string),
	}


	if strings.TrimSpace(form.Nr_faktury) == "" {
		form.FieldErrors["nr_faktury"] = "Numer faktury nie może być pusty."
	}

	if form.NIP == "" || len(form.NIP) != 10 {
		form.FieldErrors["nip"] = "NIP musi mieć 10 cyfr."
	} else if _, err := strconv.Atoi(form.NIP); err != nil {
		form.FieldErrors["nip"] = "NIP może zawierać tylko cyfry."
	}
	
	if form.Netto == 0 {
		form.FieldErrors["netto"] = "Wartość netto nie może wynosić 0."
	}
	if form.Podatek == 0 {
		form.FieldErrors["podatek"] = "Podatek nie może wynosić 0."
	} else if form.Podatek > form.Netto {
		form.FieldErrors["podatek"] = "Podatek nie może być wyższy niż wartość netto."
	}

	if strings.TrimSpace(form.Nazwa) == "" {
		form.FieldErrors["nazwa"] = "Nazwa kontrahenta nie może być pusta."
	}

	if len(form.FieldErrors) > 0 {
		tmpData := app.newTemplateData(r)
		tmpData.Form = form
		app.render(w, http.StatusUnprocessableEntity, "add_invoice.tmpl", tmpData)
		return
	}

	id, err := app.invoices.Insert(form.NIP, form.Nr_faktury, float64(form.Netto), float64(form.Podatek), form.Data, form.Inv_type, form.Nazwa)
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

func (app *application) jpkViewAll(w http.ResponseWriter, r *http.Request) {
	jpks, err := app.jpks.GetAll()
	if err != nil {
		app.serverError(w, err)
	}
	data := app.newTemplateData(r)
	data.JpkListData = jpks
	app.render(w, http.StatusOK, "jpk_files.tmpl", data)
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
	id, err := app.jpks.InsertDB(jpk, string(out))
	if err != nil {
		app.serverError(w, err)
		return
	}
	// redirect to view jpk.
	http.Redirect(w, r, fmt.Sprintf("/jpk/view/%d", id), http.StatusSeeOther)

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
