package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	mssql "github.com/microsoft/go-mssqldb"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(name, email, password, company, nip string) error {
	_, err := m.DB.Exec("INSERT INTO UserCompanies VALUES (@p1, @p2)", nip, company)
	if err != nil {
		var msSQLError *mssql.Error
		if errors.As(err, &msSQLError) {
			if msSQLError.Number == 2627 || msSQLError.Number == 2601 {
				return ErrDuplicateNip
			}
		}
		return err
	}
	stmt := "INSERT INTO Users (name, email, hashed_password, company_nip, created) values (@p1, @p2, @p3, @p4, GETDATE())"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	_, err = m.DB.Exec(stmt, name, email, hashedPassword, nip)
	if err != nil {
		var msSQLError *mssql.Error
		if errors.As(err, &msSQLError) {
			if (msSQLError.Number == 2627 || msSQLError.Number == 2601) && strings.Contains(msSQLError.Message, "users_nc_email") {
				return ErrDuplicateEmail
			}
		}
		return err
	}
	return nil
}

func (m *UserModel) Authenticate(email, password string) (int, string, error) {
	var id int
	var hashedPassword []byte
	var company_nip string

	stmt := "SELECT id, hashed_password, company_nip FROM users WHERE email = @p1"

	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword, &company_nip)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", ErrInvalidCredentials
		} else {
			return 0, "", err
		}
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, "", ErrInvalidCredentials
		} else {
			return 0, "", err
		}
	}

	return id, company_nip, err
}

func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool
	stmt := "SELECT CASE WHEN EXISTS(SELECT 1 FROM users WHERE id = @p1) THEN 1 ELSE 0 END"
	err := m.DB.QueryRow(stmt, id).Scan(&exists)
	return exists, err
}
