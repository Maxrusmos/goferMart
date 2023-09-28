package user

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"goferMart/cmd/internal/database"
	"goferMart/cmd/internal/order"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// User структура представляющая пользователя
type User struct {
	ID              int    `json:"id"`
	Login           string `json:"login"`
	Password        string `json:"password"`
	UserBalance     int
	UserWithdrawals int
}

func RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	user.UserBalance = 1000
	user.UserWithdrawals = 1000
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Проверка, что логин уникален
	if isLoginTaken(user.Login) {
		http.Error(w, "Login already taken", http.StatusConflict)
		return
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Вставка пользователя в базу данных
	session, err := database.GetSession()
	if err != nil {
		panic(err)
	}
	_, err = session.Exec("INSERT INTO gofermartUsersTable (login, password, userbalance, userwithdrawals) VALUES ($1, $2, $3, $4)", user.Login, hashedPassword, user.UserBalance, user.UserWithdrawals)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	userID, err := GetUserIDByCredentials(user.Login, hashedPassword)
	if err != nil {
		panic(err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "UserID",
		Value:    strconv.Itoa(userID),
		HttpOnly: true,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
}

func isLoginTaken(login string) bool {
	session, err := database.GetSession()
	if err != nil {
		panic(err)
	}

	var count int
	err = session.QueryRow("SELECT COUNT(*) FROM gofermartUsersTable WHERE login = $1", login).Scan(&count)
	if err != nil {
		panic(err)
	}

	return count > 0
}

func LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}
	dbUser, err := findUserByLogin(user.Login)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password)); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "UserID",
		Value:    strconv.Itoa(dbUser.ID),
		HttpOnly: true,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
}

func findUserByLogin(login string) (User, error) {
	session, err := database.GetSession()
	if err != nil {
		return User{}, err
	}

	var user User
	err = session.QueryRow("SELECT * FROM gofermartUsersTable WHERE login = $1", login).Scan(&user.ID, &user.Login, &user.Password, &user.UserBalance, &user.UserWithdrawals)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func isAuthenticated(r *http.Request) bool {
	intUserID, err := GetUserIDByCookie(r)
	if err != nil {
		return false
	}
	db, err := sql.Open("postgres", "user=postgres password=490Sutud dbname=gofermartUsers sslmode=disable")
	if err != nil {
		return false
	}
	defer db.Close()

	var dbUserID int
	err = db.QueryRow("SELECT id FROM gofermartUsersTable WHERE id = $1", intUserID).Scan(&dbUserID)
	if err != nil {
		return false
	}

	return intUserID == dbUserID
}

func UploadOrderHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	orderNumber := string(body)

	if !order.IsValidOrderNumber(orderNumber) {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	intUserID, err := GetUserIDByCookie(r)
	if err != nil {
		return
	}
	orderID, err := order.UploadOrder(orderNumber, intUserID)
	if err != nil {
		if errors.Is(err, order.ErrInvalidOrderNumber) {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		} else if errors.Is(err, order.ErrOrderAlreadyUploaded) {
			http.Error(w, err.Error(), http.StatusOK)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(fmt.Sprintf("Order uploaded successfully. Order ID: %d", orderID)))
}

func GetUserOrdersHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	intUserID, err := GetUserIDByCookie(r)
	if err != nil {
		return
	}

	orders, err := order.GetUserOrders(intUserID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func GetUserBalanceHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	intUserID, err := GetUserIDByCookie(r)
	if err != nil {
		return
	}

	var user User

	db, err := sql.Open("postgres", "user=postgres password=490Sutud dbname=gofermartUsers sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	query := "SELECT userbalance, userwithdrawals FROM gofermartUsersTable WHERE id = $1"

	errQuery := db.QueryRow(query, intUserID).Scan(&user.UserBalance, &user.UserWithdrawals)
	if errQuery != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := struct {
		Current   int `json:"current"`
		Withdrawn int `json:"withdrawn"`
	}{
		Current:   user.UserBalance,
		Withdrawn: user.UserWithdrawals,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func WithdrawBalanceHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var withdrawalRequest struct {
		OrderNumber string `json:"order"`
		Sum         int    `json:"sum"`
	}

	err := json.NewDecoder(r.Body).Decode(&withdrawalRequest)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	intUserID, err := GetUserIDByCookie(r)
	if err != nil {
		return
	}

	orderExists, err := order.IsOrderValid(intUserID, withdrawalRequest.OrderNumber)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if !orderExists {
		http.Error(w, "Unprocessable Entity", http.StatusUnprocessableEntity)
		return
	}

	var userBalance int
	db, err := sql.Open("postgres", "user=postgres password=490Sutud dbname=gofermartUsers sslmode=disable")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(
		"INSERT INTO withdrawalTable (order_number, sum, user_id) VALUES ($1, $2, $3)",
		withdrawalRequest.OrderNumber, withdrawalRequest.Sum, intUserID)
	if err != nil {
		panic(err)
	}

	query := "UPDATE orders SET sum = sum + $1 WHERE user_id = $2 AND order_number = $3"
	_, err = db.Exec(query, withdrawalRequest.Sum, intUserID, withdrawalRequest.OrderNumber)
	if err != nil {
		panic(err)
	}

	err = db.QueryRow("SELECT userbalance FROM gofermartUsersTable WHERE id = $1", intUserID).Scan(&userBalance)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if userBalance < int(withdrawalRequest.Sum) {
		http.Error(w, "Not Enough Funds", http.StatusPaymentRequired)
		return
	}

	_, err = db.Exec("UPDATE gofermartUsersTable SET userbalance = userbalance - $1 WHERE id = $2", withdrawalRequest.Sum, intUserID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type Withdrawal struct {
	OrderNumber string    `json:"order"`
	Sum         int       `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

func GetWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	intUserID, err := GetUserIDByCookie(r)
	if err != nil {
		return
	}

	db, err := sql.Open("postgres", "user=postgres password=490Sutud dbname=gofermartUsers sslmode=disable")
	if err != nil {
		panic(err)
	}
	rows, err := db.Query("SELECT order_number, sum, processed_at FROM withdrawalTable WHERE user_id = $1 ORDER BY processed_at ASC", intUserID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var withdrawals []Withdrawal

	for rows.Next() {
		var withdrawal Withdrawal
		if err := rows.Scan(&withdrawal.OrderNumber, &withdrawal.Sum, &withdrawal.ProcessedAt); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func GetOrderAccrualHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	intUserID, err := GetUserIDByCookie(r)
	if err != nil {
		return
	}
	vars := mux.Vars(r)
	orderNumber := vars["number"]

	orderExists, err := order.CheckOrderExists(orderNumber)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !orderExists {
		http.Error(w, "Order Not Found", http.StatusNotFound)
		return
	}
	accrual, status, err := MakeRequestToAccrualSystem(orderNumber, intUserID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := struct {
		Order   string `json:"order"`
		Status  string `json:"status"`
		Accrual int    `json:"accrual,omitempty"`
	}{
		Order:  orderNumber,
		Status: status,
	}

	if status == "PROCESSED" {
		response.Accrual = accrual
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
