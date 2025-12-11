package models

import (
	"database/sql"
	"errors"
	"time"
)

type Invoice struct {
	Id int
	Nr_faktury string
	Nip string
	Netto float64
	Podatek float64
	Data time.Time
	Inv_type InvoiceType
}

type InvoiceModel struct {
	DB *sql.DB
}

type InvoiceType string
const (
	SaleInvoice InvoiceType = "SALE" 
	PurchaseInvoice InvoiceType = "PURC"
)

func (m *InvoiceModel) Insert(nip string, nr_faktury string, netto float64, podatek float64, data time.Time, inv_type InvoiceType, nazwa string) (int, error) {
	stmt := "INSERT INTO Invoices (nip, nr_faktury, netto, podatek, data, type) OUTPUT Inserted.id VALUES (@p1, @p2, @p3, @p4, @p5, @p6)"
	m.DB.Exec("INSERT INTO Companies VALUES (@p1, @p2, 'PL')", nip, nazwa)
	var resId int
	err := m.DB.QueryRow(stmt, nip, nr_faktury, netto, podatek, data, inv_type).Scan(&resId)
	if err != nil {
		return 0, err
	}
	return resId, nil
}

func (m *InvoiceModel) Get(id int) (*Invoice, string, error) {
	stmt := "SELECT id, nip, nr_faktury, netto, podatek, data, type FROM Invoices WHERE id = @p1"
	cStmt := "SELECT nazwa FROM Companies WHERE nip = @p1"
	row := m.DB.QueryRow(stmt, id)
	inv := &Invoice{}
	err := row.Scan(&inv.Id, &inv.Nip, &inv.Nr_faktury, &inv.Netto, &inv.Podatek, &inv.Data, &inv.Inv_type)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", ErrNoRecord
		} else {
			return nil, "", err
		}
	}
	cRow := m.DB.QueryRow(cStmt, inv.Nip)
	var cName string
	err = cRow.Scan(&cName)
	if err != nil {
		return nil, "", err
	}
	return inv, cName, nil

}

func (m *InvoiceModel) LastMonth() ([]*Invoice, error) {
	stmt := "SELECT * FROM Invoices WHERE data >= DATEADD(month, DATEDIFF(month, 0, GETDATE()) - 1, 0) AND data < DATEADD(month, DATEDIFF(month, 0, GETDATE()), 0)"
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invoices := []*Invoice{}
	for rows.Next() {
		inv := &Invoice{}
		err := rows.Scan(&inv.Id, &inv.Nip, &inv.Nr_faktury, &inv.Netto, &inv.Podatek, &inv.Data, &inv.Inv_type)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, inv)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return invoices, err
}

func (m *InvoiceModel) GetAll() ([]*Invoice, error) {
	stmt := "SELECT * FROM Invoices WHERE data < DATEADD(month, DATEDIFF(month, 0, GETDATE()) - 1, 0) "
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	invoices := []*Invoice{}
	for rows.Next() {
		inv := &Invoice{}
		err := rows.Scan(&inv.Id, &inv.Nip, &inv.Nr_faktury, &inv.Netto, &inv.Podatek, &inv.Data, &inv.Inv_type)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, inv)
	}
	
	return invoices, nil
}

func (m *InvoiceModel) Delete(id int) (error) {
	stmt := "DELETE FROM Invoices WHERE id = @p1"
	row, err := m.DB.Exec(stmt, id)
	if err != nil {
		return err
	}
	rowsAff, err := row.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrNoRecord
	}
	return nil
}

