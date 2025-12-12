package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"app.greyhouse.es/internal/models"
	"app.greyhouse.es/internal/validator"
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
	validator.Validator
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
	data, err := time.Parse("2006-02-01",r.PostForm.Get("data"))
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
	}

	form.CheckField(validator.NotBlank(form.Nr_faktury), "nr_faktury", "Nr faktury nie może być pusty.")
	form.CheckField(validator.NotBlank(form.NIP), "nip", "NIP nie może być pusty.")
	form.CheckField(validator.NotBlank(form.Nazwa), "nazwa", "Nazwa firmy nie może być pusta.")
	form.CheckField(validator.LengthNIP(form.NIP), "nip", "NIP musi mieć 10 cyfr.")
	form.CheckField(validator.NumberNIP(form.NIP), "nip", "NIP musi składać się wyłącznie z cyfr.")
	form.CheckField(validator.NotZero(form.Netto), "netto", "Wartość netto nie może wynosić zero.")
	form.CheckField(validator.NotZero(form.Podatek), "podatek", "Wartość podatku nie może wynosić zero.")

	if !form.Valid() {
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
	// add current date to the newTemplateData constructor, check against it whether to display the delete button. 
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	inv, cname, err := app.invoices.Get(id)
	data := app.newTemplateData(r)
	data.Invoice = inv
	data.InvDeletable = inv.IsPreviousMonth()
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

func (app *application) viewAllInvoices(w http.ResponseWriter, r *http.Request) {
	invoices, err := app.invoices.GetAll()
	if err != nil {
		app.serverError(w, err)
	}
	data := app.newTemplateData(r)
	data.Invoices = invoices
	app.render(w, http.StatusOK, "view_invoices.tmpl", data)
}

func (app *application) deleteInvoice(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.serverError(w, err)
	}
	err = app.invoices.Delete(id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
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
	slices.Reverse(jpks)
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
	http.Redirect(w, r, "/jpk/viewall", http.StatusSeeOther)
}

func (app *application) jpkConfirm(w http.ResponseWriter, r *http.Request) {
	// prototype for a function that will prompt for UPO and confirm the file
	// if it's provided.
}
