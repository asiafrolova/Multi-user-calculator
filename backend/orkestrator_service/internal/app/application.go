package application

import (
	"fmt"
	"net/http"
	"os"

	"github.com/asiafrolova/Final_task/orkestrator_service/internal/database"
	handlers "github.com/asiafrolova/Final_task/orkestrator_service/internal/handlers"
	logger "github.com/asiafrolova/Final_task/orkestrator_service/internal/logger"
	"github.com/asiafrolova/Final_task/orkestrator_service/internal/repo"
	"github.com/asiafrolova/Final_task/orkestrator_service/pkg/orkestrator"
)

type Config struct {
	Addr string
}

func ConfigFromEnv() *Config {
	config := new(Config)
	config.Addr = os.Getenv("PORT")
	if config.Addr == "" {
		config.Addr = "8080"
	}
	return config
}

type Application struct {
	config *Config
}

func New() *Application {
	database.Init()
	logger.Init()
	orkestrator.InitOrkestrator()
	repo.Init()
	return &Application{config: ConfigFromEnv()}

}

func (a *Application) RunServer() error {

	logger.Info(fmt.Sprintf("Server started address: %s", a.config.Addr))
	http.HandleFunc("/api/v1/calculate", handlers.AddExpressionsHandler)
	http.HandleFunc("/api/v1/expressions", handlers.GetExpressionsListHandler)
	http.HandleFunc("/api/v1/expressions/{id}", handlers.GetExpressionByIDHandler)
	http.HandleFunc("/api/v1/register", handlers.RegisterUser)
	http.HandleFunc("/api/v1/login", handlers.LoginUser)
	http.HandleFunc("/api/v1/delete/user", handlers.DeleteUser)
	http.HandleFunc("/internal/task", handlers.GetTaskHandler)

	http.HandleFunc("/api/v1/delete/expressions", handlers.DeleteExpressions)
	http.HandleFunc("/api/v1/delete/expressions/{id}", handlers.DeleteExpressionByID)
	http.HandleFunc("/api/v1/update/user", handlers.UpdateUser)
	return http.ListenAndServe(":"+a.config.Addr, nil)
}
