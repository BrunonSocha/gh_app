package main

import (
	"bytes"
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
	NIP        string
	Nazwa      string
	Netto      float64
	Podatek    float64
	Data       time.Time
	Inv_type   models.InvoiceType
	validator.Validator
}

type confirmJpkForm struct {
	UPO string
	validator.Validator
}

type userSignupForm struct {
	Name     string
	Email    string
	Password string
	Company  string
	Nip      string
	validator.Validator
}

type userLoginForm struct {
	Email    string
	Password string
	validator.Validator
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	company_nip := app.getNIP(r)
	data := app.newTemplateData(r)
	invoices, err := app.invoices.GetAll(company_nip, data.CurrentDate)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data.Invoices = invoices
	app.sessionManager.Put(r.Context(), "date", data.CurrentDate)
	app.render(w, http.StatusOK, "home.tmpl", data)
}

func (app *application) homePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	dateStr := r.PostForm.Get("month")
	date, err := time.Parse("2006-01", dateStr)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	company_nip := app.getNIP(r)
	data := app.newTemplateData(r)
	data.CurrentDate = date
	invoices, err := app.invoices.GetAll(company_nip, data.CurrentDate)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data.Invoices = invoices
	app.sessionManager.Put(r.Context(), "date", data.CurrentDate)

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
	data, err := time.Parse("2006-01-02", r.PostForm.Get("data"))
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
	form := addInvoiceForm{
		Nr_faktury: r.PostForm.Get("nr_faktury"),
		NIP:        strings.ReplaceAll(strings.ReplaceAll(r.PostForm.Get("nip"), "-", ""), " ", ""),
		Netto:      netto,
		Podatek:    podatek,
		Data:       data,
		Nazwa:      nazwa,
		Inv_type:   inv_type,
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
	company_nip := app.getNIP(r)
	id, err := app.invoices.Insert(form.NIP, form.Nr_faktury, float64(form.Netto), float64(form.Podatek), form.Data, form.Inv_type, form.Nazwa, company_nip)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Dodano fakturę.")

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
	company_nip := app.getNIP(r)
	inv, cname, err := app.invoices.Get(id, company_nip)
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

func (app *application) deleteInvoice(w http.ResponseWriter, r *http.Request) {
	company_nip := app.getNIP(r)
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.notFound(w)
		return
	}
	err = app.invoices.Delete(id, company_nip)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Usunięto fakturę.")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) viewJpk(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.notFound(w)
		return
	}
	company_nip := app.getNIP(r)
	data := app.newTemplateData(r)
	data.Jpk, data.JpkMetadata, err = app.jpks.Get(id, company_nip)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.render(w, http.StatusOK, "view_jpk.tmpl", data)
}

func (app *application) viewAllJpk(w http.ResponseWriter, r *http.Request) {
	company_nip := app.getNIP(r)
	jpks, err := app.jpks.GetAll(company_nip)
	if err != nil {
		app.serverError(w, err)
	}
	slices.Reverse(jpks)
	data := app.newTemplateData(r)
	data.JpkListData = jpks
	app.render(w, http.StatusOK, "jpk_files.tmpl", data)
}

func (app *application) addJpk(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	company_nip := app.getNIP(r)
	data := app.newTemplateData(r)
	date, err := time.Parse("2006-01", params.ByName("date"))
	data.CurrentDate = date
	invoices, err := app.invoices.GetAll(company_nip, data.CurrentDate)
	if err != nil {
		app.serverError(w, err)
		return
	}

	jpk, err := app.jpks.NewJpk(invoices, data.CurrentDate)

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
	id, err := app.jpks.InsertDB(jpk, string(out), company_nip)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Wygenerowano JPK.")

	// redirect to view jpk.
	http.Redirect(w, r, fmt.Sprintf("/jpk/view/%d", id), http.StatusSeeOther)

}

func (app *application) deleteJpk(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.notFound(w)
		return
	}
	company_nip := app.getNIP(r)
	err = app.jpks.Delete(id, company_nip)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Usunięto wersję roboczą JPK.")

	http.Redirect(w, r, "/jpk/viewall", http.StatusSeeOther)
}

func (app *application) downloadJpk(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.serverError(w, err)
		return
	}
	company_nip := app.getNIP(r)
	fileContent, err := app.jpks.GetContent(id, company_nip)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"jpk_v7m_%d.xml\"", id))
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Length", strconv.Itoa(len(fileContent)))

	http.ServeContent(w, r, "jpk.xml", time.Now(), bytes.NewReader(fileContent))
}

func (app *application) confirmJpk(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	company_nip := app.getNIP(r)
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.serverError(w, err)
		return
	}
	err = r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form := confirmJpkForm{
		UPO: r.PostForm.Get("upo"),
	}
	form.CheckField(validator.NotBlank(form.UPO), "upo", "UPO nie może być puste.")
	if !form.Valid() {
		jpk, metadata, err := app.jpks.Get(id, company_nip)
		if err != nil {
			app.serverError(w, err)
			return
		}
		data := app.newTemplateData(r)
		data.Jpk = jpk
		data.JpkMetadata = metadata
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "view_jpk.tmpl", data)
		return
	}
	err = app.jpks.Confirm(id, form.UPO, company_nip)

	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	app.sessionManager.Put(r.Context(), "flash", "Zatwierdzono JPK.")
	http.Redirect(w, r, fmt.Sprintf("/jpk/view/%d", id), http.StatusSeeOther)
}

func (app *application) userSignUp(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}
	app.render(w, http.StatusOK, "signup.tmpl", data)
}

func (app *application) userSignUpPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form := userSignupForm{
		Name:     r.PostForm.Get("name"),
		Email:    r.PostForm.Get("email"),
		Password: r.PostForm.Get("password"),
		Company:  r.PostForm.Get("company"),
		Nip:      r.PostForm.Get("nip"),
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "Nazwa nie może być pusta")
	form.CheckField(validator.NotBlank(form.Email), "email", "Email nie może być pusty")
	form.CheckField(validator.NotBlank(form.Password), "password", "Hasło nie może być puste")
	form.CheckField(validator.NotBlank(form.Nip), "nip", "NIP nie może być pusty")
	form.CheckField(validator.NotBlank(form.Company), "company", "Nazwa firmy nie może być pusta")
	form.CheckField(validator.Matches(form.Email, validator.EmailRegex), "email", "Email musi być poprawny")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "Hasło musi mieć min. 8 znaków")
	form.CheckField(validator.LengthNIP(form.Nip), "nip", "NIP musi być poprawny")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
		return
	}

	err = app.users.Insert(form.Name, form.Email, form.Password, form.Company, form.Nip)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
		} else if errors.Is(err, models.ErrDuplicateNip) {
			form.AddFieldError("nip", "NIP is already in use")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
		} else {
			app.serverError(w, err)
		}

		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Zarejestrowano. Zaloguj się")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}

	app.render(w, http.StatusOK, "login.tmpl", data)

}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := userLoginForm{
		Email:    r.PostForm.Get("email"),
		Password: r.PostForm.Get("password"),
	}
	form.CheckField(validator.NotBlank(form.Email), "email", "Wprowadź email.")
	form.CheckField(validator.NotBlank(form.Password), "password", "Wprowadź hasło.")
	form.CheckField(validator.Matches(form.Email, validator.EmailRegex), "email", "Wprowadź poprawny email.")
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "login.tmpl", data)
		return
	}

	id, nip, err := app.users.Authenticate(form.Email, form.Password)

	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "login.tmpl", data)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// good practice to renew session after changing privileges
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)
	app.sessionManager.Put(r.Context(), "authenticatedUserNIP", nip)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Remove(r.Context(), "authenticatedUserNIP")
	app.sessionManager.Put(r.Context(), "flash", "Wylogowano.")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
