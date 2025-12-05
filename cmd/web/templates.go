package main

import "app.greyhouse.es/internal/models"

type templateData struct {
	Invoice *models.Invoice
	Invoices []*models.Invoice
	CompanyName string
	TotalVat float64
}
