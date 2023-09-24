// db/db.go

package database

import (
	"database/sql"

	_ "github.com/lib/pq"
)

// DB - это объект соединения с базой данных
type DB struct {
	*sql.DB
}

var db *DB

// InitDB инициализирует соединение с базой данных
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

	// SQL-запрос для создания таблицы, если она не существует
	query := `
    CREATE TABLE IF NOT EXISTS gofermartUsersTable (
        id SERIAL PRIMARY KEY,
        login TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL
    )
    `

	_, err = session.Exec(query)
	return err
}

func GetSession() (*sql.DB, error) {
	return db.DB, nil
}
