package application

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/asiafrolova/Multi-user-calculator/orkestrator_service/internal/database"
	handlers "github.com/asiafrolova/Multi-user-calculator/orkestrator_service/internal/handlers"
	logger "github.com/asiafrolova/Multi-user-calculator/orkestrator_service/internal/logger"
	"github.com/asiafrolova/Multi-user-calculator/orkestrator_service/internal/repo"
	"github.com/asiafrolova/Multi-user-calculator/orkestrator_service/pkg/orkestrator"
	"google.golang.org/grpc"

	"time"

	pb "github.com/asiafrolova/Multi-user-calculator/proto"
)

var (
	WAITING_TIME = time.Millisecond * 100
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
	pb.CalculatorServiceServer
}

func New() *Application {
	database.Init()
	logger.Init()
	orkestrator.InitOrkestrator()
	repo.Init()
	return &Application{config: ConfigFromEnv()}

}
func (a *Application) RunServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", "5000"))
	if err != nil {
		logger.Error(fmt.Sprintf("error starting tcp listener: %v", err))

	}
	logger.Info(fmt.Sprintf("tcp server started at:%s", "5000"))
	grpcServer := grpc.NewServer()
	pb.RegisterCalculatorServiceServer(grpcServer, a)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error(fmt.Sprintf("error serving grpc:%v ", err))

		}
	}()

	mux := http.NewServeMux()

	// Регистрируем маршруты
	mux.Handle("/api/v1/calculate", http.HandlerFunc(handlers.AddExpressionsHandler))
	mux.Handle("/api/v1/expressions", http.HandlerFunc(handlers.GetExpressionsListHandler))
	mux.Handle("/api/v1/delete/expressions/{id}", http.HandlerFunc(handlers.DeleteExpressionByID))
	mux.Handle("/api/v1/expressions/{id}", http.HandlerFunc(handlers.GetExpressionByIDHandler))
	mux.Handle("/api/v1/register", http.HandlerFunc(handlers.RegisterUser))
	mux.Handle("/api/v1/login", http.HandlerFunc(handlers.LoginUser))
	mux.Handle("/api/v1/delete/user", http.HandlerFunc(handlers.DeleteUser))

	mux.Handle("/api/v1/delete/expressions", http.HandlerFunc(handlers.DeleteExpressions))
	mux.Handle("/api/v1/update/user", http.HandlerFunc(handlers.UpdateUser))

	// Оборачиваем весь mux в CORS middleware
	handler := handlers.WithCORS(mux)
	logger.Info(fmt.Sprintf("Server started at :%s", a.config.Addr))
	for {
		http.ListenAndServe(":"+a.config.Addr, handler)
	}

}
func (a *Application) GetTask(ctx context.Context, in *pb.GetTaskRequest) (*pb.GetTaskResponse, error) {
	timer := time.NewTimer(WAITING_TIME)
	repo.Init()
	repo.GetSimpleOperations()
	select {
	case <-timer.C:
		//Если время истекло отвечаем, что нет задач
		return nil, orkestrator.ErrNotExpression
	case sExp := <-repo.SimpleExpressions:
		//Задача нашлась
		response := pb.GetTaskResponse{
			Id:            sExp.Id,
			Arg1:          sExp.Arg1,
			Arg2:          sExp.Arg2,
			Operation:     sExp.Operation,
			OperationTime: float32(sExp.Operation_time),
		}

		logger.Info(fmt.Sprintf("Task %v sent successfully", sExp))
		return &response, nil
	}
}
func (a *Application) PushResult(ctx context.Context, in *pb.PushResultRequest) (*pb.PushResultResponse, error) {

	if in.Error != "" {
		//Пришла ошибка при вычислении
		err := repo.SetResult(in.Id, 0, fmt.Errorf(in.Error))
		if err != nil {
			return nil, err
		} else {
			logger.Info("Got an error in the task")
		}
		return nil, nil

	} else {

		err := repo.SetResult(in.Id, float64(in.Result), nil)
		if err != nil {
			return nil, err

		} else {
			logger.Info("The task result was successfully written")
			return nil, nil
		}

	}

}
