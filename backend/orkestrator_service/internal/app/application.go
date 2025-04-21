package application

import (
	"fmt"
	"net/http"
	"os"

	"github.com/asiafrolova/Multi-user-calculator/orkestrator_service/internal/database"
	handlers "github.com/asiafrolova/Multi-user-calculator/orkestrator_service/internal/handlers"
	logger "github.com/asiafrolova/Multi-user-calculator/orkestrator_service/internal/logger"
	"github.com/asiafrolova/Multi-user-calculator/orkestrator_service/internal/repo"
	"github.com/asiafrolova/Multi-user-calculator/orkestrator_service/pkg/orkestrator"
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
func (a *Application) RunServer() {
	mux := http.NewServeMux()

	// Регистрируем маршруты
	mux.Handle("/api/v1/calculate", http.HandlerFunc(handlers.AddExpressionsHandler))
	mux.Handle("/api/v1/expressions", http.HandlerFunc(handlers.GetExpressionsListHandler))
	mux.Handle("/api/v1/delete/expressions/{id}", http.HandlerFunc(handlers.DeleteExpressionByID))
	mux.Handle("/api/v1/expressions/{id}", http.HandlerFunc(handlers.GetExpressionByIDHandler))
	mux.Handle("/api/v1/register", http.HandlerFunc(handlers.RegisterUser))
	mux.Handle("/api/v1/login", http.HandlerFunc(handlers.LoginUser))
	mux.Handle("/api/v1/delete/user", http.HandlerFunc(handlers.DeleteUser))
	mux.Handle("/internal/task", http.HandlerFunc(handlers.GetTaskHandler))
	mux.Handle("/api/v1/delete/expressions", http.HandlerFunc(handlers.DeleteExpressions))
	mux.Handle("/api/v1/update/user", http.HandlerFunc(handlers.UpdateUser))

	// Оборачиваем весь mux в CORS middleware
	handler := handlers.WithCORS(mux)
	logger.Info(fmt.Sprintf("Server started at :%s", a.config.Addr))
	for {
		http.ListenAndServe(":"+a.config.Addr, handler)
	}

}
