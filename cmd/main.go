package main

import (
	"fmt"
	"net/http"

	"goferMart/cmd/internal/database"
	"goferMart/cmd/internal/order"
	"goferMart/cmd/internal/user"

	"github.com/gorilla/mux"
)

func main() {
	// Инициализация базы данных
	if err := database.InitDB("user=postgres password=490Sutud dbname=gofermartUsers sslmode=disable"); err != nil {
		panic(err)
	}
	if err := database.CreateGofermartUsersTable(); err != nil {
		panic(err)
	}
	if err := order.CreateTableIfNotExists(); err != nil {
		panic(err)
	}

	r := mux.NewRouter()

	// Регистрация хендлеров
	r.HandleFunc("/api/user/register", user.RegisterUserHandler).Methods("POST")
	r.HandleFunc("/api/user/login", user.LoginUserHandler).Methods("POST")    // Настройка маршрута для аутентификации
	r.HandleFunc("/api/user/orders", user.UploadOrderHandler).Methods("POST") // Обновленный маршрут для загрузки заказов

	http.Handle("/", r)

	fmt.Println("Server is listening on port 8080...")
	http.ListenAndServe(":8080", nil)
}
