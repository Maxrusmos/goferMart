package user

import (
	"database/sql"
	"net/http"
	"strconv"
)

func GetUserIDByCookie(r *http.Request) (int, error) {
	cookie, err := r.Cookie("UserID")
	if err != nil {
		return 0, err
	}

	userID := cookie.Value
	intUserID, err := strconv.Atoi(userID)
	if err != nil {
		return 0, err
	}
	return intUserID, nil
}

func GetUserIDByCredentials(login string, password []byte) (int, error) {
	db, err := sql.Open("postgres", "user=postgres password=490Sutud dbname=gofermartUsers sslmode=disable")
	if err != nil {
		return 0, err
	}
	defer db.Close()
	query := "SELECT id FROM gofermartUsersTable WHERE login = $1 AND password = $2"
	var userID int
	err = db.QueryRow(query, login, password).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func MakeRequestToAccrualSystem(orderNumber string, userID int) (int, string, error) {
	db, err := sql.Open("postgres", "user=postgres password=490Sutud dbname=gofermartUsers sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	query := "SELECT sum FROM orders WHERE user_id = $1 AND order_number = $2"
	var sum int
	err = db.QueryRow(query, userID, orderNumber).Scan(&sum)
	if err != nil {
		panic(err)
	}
	accrual := 0
	if sum > 500 {
		accrual = 200
	}
	return accrual, "PROCESSED", nil
}
