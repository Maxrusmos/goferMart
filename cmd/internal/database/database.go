// db/db.go

package database

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

var db *DB

func InitDB(dbConnStr string) error {
	database, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		return err
	}
	db = &DB{database}
	return nil
}

func CreateGofermartUsersTable() error {
	session, err := GetSession()
	if err != nil {
		return err
	}
	query := `
    CREATE TABLE IF NOT EXISTS gofermartUsersTable (
        id SERIAL PRIMARY KEY,
        login TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL,
		userBalance INT,
		userWithdrawals INT
    )
    `

	_, err = session.Exec(query)
	return err
}

func CreateOrdersTableIfNotExists() error {
	session, err := GetSession()
	if err != nil {
		return err
	}
	query := `
    CREATE TABLE IF NOT EXISTS orders (
        id SERIAL PRIMARY KEY,
        order_number TEXT NOT NULL,
        user_id INT NOT NULL,
        status TEXT,
        accrual INT,
		sum INT,
        uploaded_at TIMESTAMPTZ DEFAULT NOW()
    )
    `

	_, err = session.Exec(query)
	return err
}

func CreateWithdrawalTableIfNotExists() error {
	session, err := GetSession()
	if err != nil {
		return err
	}
	query := `
        CREATE TABLE IF NOT EXISTS withdrawalTable (
            id SERIAL PRIMARY KEY,
			user_id INT NOT NULL,
            order_number TEXT NOT NULL,
            sum INT NOT NULL,
            processed_at TIMESTAMPTZ DEFAULT NOW()
        )
    `

	_, err = session.Exec(query)
	return err
}

func GetSession() (*sql.DB, error) {
	return db.DB, nil
}
