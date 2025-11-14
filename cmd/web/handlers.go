package main

import (
	"fmt"
	"net/http"
	"strconv"
	"log"
	"html/template"
)

func home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	files := []string{"./ui/html/pages/home.html", "./ui/html/base.html", "./ui/html/partials/nav.html"}


	templateset, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Wewnętrzny błąd serwera", http.StatusInternalServerError)
		return
	}
	
	err = templateset.ExecuteTemplate(w, "base", nil)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Wewnętrzny błąd serwera", http.StatusInternalServerError)
	}
}

func jpkAddInvoice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Metoda nieakceptowana", http.StatusMethodNotAllowed)
		return
	}
	w.Write([]byte("Dodaj fakturę..."))
}

func jpkView(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Metoda niekaceptowana", http.StatusMethodNotAllowed)
	}
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	fmt.Fprintf(w, "Wyświetl JPK o ID %d...", id)
}

func jpkCreate(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Tworzenie pliku JPK..."))
}
