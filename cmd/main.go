package main

import (
	"flag"
	"fmt"
	"net/http"

	"goferMart/cmd/internal/database"
	"goferMart/cmd/internal/user"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

func main() {
	var (
		runAddressFlag           = flag.String("a", "localhost:8080", "Address and port to run the service")
		databaseURIFlag          = flag.String("d", "user=postgres password=490Sutud dbname=gofermartUsers sslmode=disable", "Database connection URI")
		accrualSystemAddressFlag = flag.String("r", "", "Address of the accrual system")
	)

	flag.Parse()

	if *runAddressFlag != "" {
		viper.Set("run_address", *runAddressFlag)
	}
	if *databaseURIFlag != "" {
		viper.Set("database_uri", *databaseURIFlag)
	}
	if *accrualSystemAddressFlag != "" {
		viper.Set("accrual_system_address", *accrualSystemAddressFlag)
	}

	viper.SetEnvPrefix("LOYALTY")
	viper.AutomaticEnv()

	viper.SetDefault("run_address", ":8080")
	viper.SetDefault("database_uri", "postgres://postres:490Sutud@localhost:5432/gofermartUsers")
	viper.SetDefault("accrual_system_address", "http://localhost:8000")

	runAddress := viper.GetString("run_address")
	databaseURI := viper.GetString("database_uri")
	// accrualSystemAddress := viper.GetString("accrual_system_address")

	if err := database.InitDB(databaseURI); err != nil {
		panic(err)
	}
	if err := database.CreateGofermartUsersTable(); err != nil {
		panic(err)
	}
	if err := database.CreateOrdersTableIfNotExists(); err != nil {
		panic(err)
	}
	if err := database.CreateWithdrawalTableIfNotExists(); err != nil {
		panic(err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/api/user/register", user.RegisterUserHandler).Methods("POST")
	r.HandleFunc("/api/user/login", user.LoginUserHandler).Methods("POST")
	r.HandleFunc("/api/user/orders", user.UploadOrderHandler).Methods("POST")
	r.HandleFunc("/api/user/orders", user.GetUserOrdersHandler).Methods("GET")
	r.HandleFunc("/api/user/balance", user.GetUserBalanceHandler).Methods("GET")
	r.HandleFunc("/api/user/balance/withdraw", user.WithdrawBalanceHandler).Methods("POST")
	r.HandleFunc("/api/user/withdrawals", user.GetWithdrawalsHandler).Methods("GET")
	r.HandleFunc("/api/orders/{number}", user.GetOrderAccrualHandler).Methods("GET")

	http.Handle("/", r)

	fmt.Println("Server is listening on port 8080...")
	http.ListenAndServe(runAddress, nil)
}
