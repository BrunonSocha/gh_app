package main

import (
	"fmt"
	"net/http"
	"strconv"
)

func home(w http.ResponseWriter, r *http.Request) {
	// sprawdzamy czy path to dokładnie "/". 
	// inaczej rzucamy 404
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Write([]byte("Cześć."))
}

func jpkAddInvoice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// WriteHeader można zawołać tylko raz na odpowiedź
		// jeśli nie wywołamy w.WriteHeader() to w.Write
		// automatycznie zwróci 200 OK
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
