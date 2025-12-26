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
			if msSQLError.Number == 1062 && strings.Contains(msSQLError.Message, "users_nc_nip") {
				return ErrDuplicateNip
			}
		}
		return err
	}
	stmt := "INSERT INTO Users (name, email, hashed_password, nip, created) values (@p1, @p2, @p3, @p4, GETDATE())"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	_, err = m.DB.Exec(stmt, name, email, hashedPassword, nip)
	if err != nil {
		var msSQLError *mssql.Error
		if errors.As(err, &msSQLError) {
			if msSQLError.Number == 1062 && strings.Contains(msSQLError.Message, "users_nc_email") {
				return ErrDuplicateEmail
			}
		}
		return err
	}
	return nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte

	stmt := "SELECT id, hashed_password FROM users WHERE email = @p1"

	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	return id, err
}

func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool
	stmt := "SELECT CASE WHEN EXISTS(SELECT 1 FROM users WHERE id = @p1) THEN 1 ELSE 0 END"
	err := m.DB.QueryRow(stmt, id).Scan(&exists)
	return exists, err
}
