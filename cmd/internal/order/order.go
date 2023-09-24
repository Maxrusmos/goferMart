// order/order.go

package order

import (
	"database/sql"
	"errors"
	"fmt"
	"goferMart/cmd/internal/database"
)

var ErrInvalidOrderNumber = errors.New("Invalid order number format")
var ErrOrderAlreadyUploaded = errors.New("Order already uploaded by this user")

// Order представляет информацию о заказе
type Order struct {
	ID          int
	OrderNumber string
	UserID      int
}

// UploadOrder загружает номер заказа в базу данных
func UploadOrder(orderNumber string, userID int) (int, error) {
	// Проверьте, существует ли заказ с таким номером в базе данных
	exists, err := checkOrderExists(orderNumber)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, ErrOrderAlreadyUploaded
	}

	// Добавьте заказ в базу данных, связывая его с userID
	orderID, err := insertOrder(orderNumber, userID)
	if err != nil {
		return 0, err
	}

	return orderID, nil
}

func checkOrderExists(orderNumber string) (bool, error) {
	// Ваша строка подключения к базе данных PostgreSQL
	db, err := sql.Open("postgres", "user=postgres password=490Sutud dbname=gofermartUsers sslmode=disable")
	if err != nil {
		return false, err
	}
	defer db.Close()

	// Выполнение SQL-запроса для проверки существования заказа
	query := "SELECT EXISTS (SELECT 1 FROM orders WHERE order_number = $1)"
	var exists bool
	err = db.QueryRow(query, orderNumber).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func IsValidOrderNumber(orderNumber string) bool {
	fmt.Println(orderNumber)
	// Здесь вы можете реализовать логику проверки корректности номера заказа
	// Например, проверка на формат или алгоритм Луна
	return true
}

func isOrderUploadedByUser(orderNumber string, userID int) (bool, error) {
	session, err := database.GetSession()
	if err != nil {
		return false, err
	}

	var count int
	err = session.QueryRow("SELECT COUNT(*) FROM orders WHERE order_number = $1 AND user_id = $2", orderNumber, userID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// CreateTableIfNotExists создает таблицу orders, если она не существует
func CreateTableIfNotExists() error {
	session, err := database.GetSession()
	if err != nil {
		return err
	}

	// SQL-запрос для создания таблицы, если она не существует
	query := `
    CREATE TABLE IF NOT EXISTS orders (
        id SERIAL PRIMARY KEY,
        order_number TEXT NOT NULL,
        user_id INT NOT NULL
    )
    `

	_, err = session.Exec(query)
	return err
}

func insertOrder(orderNumber string, userID int) (int, error) {
	// Ваша строка подключения к базе данных PostgreSQL
	fmt.Println("userID", userID)
	db, err := sql.Open("postgres", "user=postgres password=490Sutud dbname=gofermartUsers sslmode=disable")
	if err != nil {
		return 0, err
	}
	defer db.Close()
	query := "INSERT INTO orders (order_number, user_id) VALUES ($1, $2) RETURNING id"
	var orderID int
	err = db.QueryRow(query, orderNumber, userID).Scan(&orderID)
	if err != nil {
		return 0, err
	}

	return orderID, nil
}
