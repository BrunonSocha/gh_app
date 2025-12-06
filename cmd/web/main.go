package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"app.greyhouse.es/internal/models"
	_ "github.com/microsoft/go-mssqldb"
)

type application struct {
	errorLog *log.Logger
	infoLog *log.Logger
	invoices *models.InvoiceModel
	invTypes *models.InvoiceType
	templateCache map[string]*template.Template
}

func main() {
	// command line flag for the port
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn_str := fmt.Sprintf("sqlserver://sa:%s@localhost:1433?database=Greyhouse&trustServerCertificate=true", url.QueryEscape("DavidBowie11%"))
	dsn := flag.String("dsn", dsn_str, "greyhouse-sql")
	flag.Parse()

	// creating loggers
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	
	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	// application struct instance
	app := &application{
		errorLog: errorLog,
		infoLog: infoLog,
		invoices: &models.InvoiceModel{DB: db},
		templateCache: templateCache,
	}

	// our own server
	srv := &http.Server{
		Addr: *addr,
		ErrorLog: errorLog,
		Handler: app.routes(),
	}

	// on run
	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

type neuteredFileSystem struct {
	fs http.FileSystem
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}
	return f, nil
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlserver", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
