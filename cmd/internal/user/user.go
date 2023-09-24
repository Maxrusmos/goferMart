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

	"golang.org/x/crypto/bcrypt"
)

// User структура представляющая пользователя
type User struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

func RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
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
	_, err = session.Exec("INSERT INTO gofermartUsersTable (login, password) VALUES ($1, $2)", user.Login, hashedPassword)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	userID, err := getUserIDByCredentials(user.Login, hashedPassword)
	if err != nil {
		panic(err)
	}

	// Аутентификация пользователя (просто для примера, вы можете использовать свой механизм)
	// В данном случае мы просто устанавливаем куки с идентификатором пользователя
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

	// Поиск пользователя по логину
	dbUser, err := findUserByLogin(user.Login)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password)); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Успешная аутентификация
	w.WriteHeader(http.StatusOK)
}

func findUserByLogin(login string) (User, error) {
	session, err := database.GetSession()
	if err != nil {
		return User{}, err
	}

	var user User
	err = session.QueryRow("SELECT * FROM gofermartUsersTable WHERE login = $1", login).Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func isAuthenticated(r *http.Request) bool {
	// Получите куки из запроса
	cookie, err := r.Cookie("UserID")
	if err != nil {
		return false // Куки "UserID" отсутствует, пользователь не аутентифицирован
	}
	fmt.Println("COOKIE====", cookie)

	// Дополнительные проверки аутентификации могут быть добавлены здесь
	// Например, вы можете проверить, что значение куки соответствует действительному пользователю в вашей системе

	return true // Куки "UserID" найден, пользователь аутентифицирован
}

func getUserIDByCredentials(login string, password []byte) (int, error) {
	// Ваша строка подключения к базе данных PostgreSQL
	db, err := sql.Open("postgres", "user=postgres password=490Sutud dbname=gofermartUsers sslmode=disable")
	if err != nil {
		return 0, err
	}
	defer db.Close()

	// Выполнение SQL-запроса для извлечения id по логину и паролю
	query := "SELECT id FROM gofermartUsersTable WHERE login = $1 AND password = $2"
	var userID int
	err = db.QueryRow(query, login, password).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func userIDFromContext(r *http.Request) int {
	userID, ok := r.Context().Value("userID").(int)
	fmt.Println("context", userID)
	if !ok {
		// Обработка случая, если userID не найден в контексте
		// Возможно, здесь следует вернуть ошибку или другое значение по умолчанию
		return 0 // Вернуть значение по умолчанию (например, 0) или выполнить другую обработку
	}
	return userID
}

func UploadOrderHandler(w http.ResponseWriter, r *http.Request) {
	// Проверка аутентификации пользователя (псевдокод)
	if !isAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Прочитайте тело запроса и получите номер заказа
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	orderNumber := string(body)

	// Проверка наличия корректного номера заказа
	if !order.IsValidOrderNumber(orderNumber) {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	// Загрузка номера заказа
	cookie, err := r.Cookie("UserID")
	if err != nil {
		// Обработка ошибки, если кука не найдена
		http.Error(w, "Cookie not found", http.StatusUnauthorized)
		return
	}

	// Получение значения куки
	userID := cookie.Value
	fmt.Println("userID", userID)
	intUserID, err := strconv.Atoi(userID)
	if err != nil {
		fmt.Println("Ошибка:", err)
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

	// Возвращение успешного статуса и ID загруженного заказа
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(fmt.Sprintf("Order uploaded successfully. Order ID: %d", orderID)))
}
