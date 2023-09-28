// order/order.go

package order

import (
	"database/sql"
	"errors"
	"fmt"
	"goferMart/cmd/internal/database"
	"time"
)

var ErrInvalidOrderNumber = errors.New("Invalid order number format")
var ErrOrderAlreadyUploaded = errors.New("Order already uploaded by this user")

type Order struct {
	ID          int
	OrderNumber string `json:"number"`
	UserID      int
	Status      string    `json:"status"`
	Accrual     int       `json:"accrual,omitempty"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

func UploadOrder(orderNumber string, userID int) (int, error) {
	exists, err := CheckOrderExists(orderNumber)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, ErrOrderAlreadyUploaded
	}

	orderID, err := insertOrder(orderNumber, userID)
	if err != nil {
		return 0, err
	}

	return orderID, nil
}

func CheckOrderExists(orderNumber string) (bool, error) {
	db, err := sql.Open("postgres", "user=postgres password=490Sutud dbname=gofermartUsers sslmode=disable")
	if err != nil {
		return false, err
	}
	defer db.Close()
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

func insertOrder(orderNumber string, userID int) (int, error) {
	fmt.Println("userID", userID)
	db, err := sql.Open("postgres", "user=postgres password=490Sutud dbname=gofermartUsers sslmode=disable")
	if err != nil {
		return 0, err
	}
	defer db.Close()
	query := "INSERT INTO orders (order_number, user_id, status, accrual, sum) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	var orderID int
	err = db.QueryRow(query, orderNumber, userID, "NEW", 0, 0).Scan(&orderID)
	if err != nil {
		return 0, err
	}

	return orderID, nil
}

type OrderResponse struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	UploadedAt time.Time `json:"uploaded_at"`
	Accrual    int       `json:"accrual,omitempty"`
}

func GetUserOrders(userID int) ([]OrderResponse, error) {
	db, err := sql.Open("postgres", "user=postgres password=490Sutud dbname=gofermartUsers sslmode=disable")
	query := "SELECT order_number, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at ASC"
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []OrderResponse

	for rows.Next() {
		var order OrderResponse
		if err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func IsOrderValid(userID int, orderNumber string) (bool, error) {
	db, err := sql.Open("postgres", "user=postgres password=490Sutud dbname=gofermartUsers sslmode=disable")
	if err != nil {
		return false, err
	}
	defer db.Close()

	query := "SELECT COUNT(*) FROM orders WHERE user_id = $1 AND order_number = $2"
	var count int
	err = db.QueryRow(query, userID, orderNumber).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
