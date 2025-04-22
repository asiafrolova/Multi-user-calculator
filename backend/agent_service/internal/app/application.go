package application

import (
	"context"

	"fmt"
	"os"
	"strconv"
	"time"

	logger "github.com/asiafrolova/Multi-user-calculator/agent_service/internal/logger"
	calculator "github.com/asiafrolova/Multi-user-calculator/agent_service/pkg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/asiafrolova/Multi-user-calculator/proto"
)

type Config struct {
	Addr            string
	COMPUTING_POWER int
}

func ConfigFromEnv() *Config {
	config := new(Config)
	config.Addr = os.Getenv("PORT")
	var err error
	config.COMPUTING_POWER, err = strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if config.Addr == "" {
		config.Addr = "8080"
	}
	if err != nil || config.COMPUTING_POWER == 0 {
		config.COMPUTING_POWER = 5
	}
	return config
}

type Application struct {
	config *Config
}
type RequestResult struct {
	Id     string  `json:"id"`
	Result float64 `json:"result"`
	Err    string  `json:"error"`
}
type ResponseTask struct {
	Task calculator.SimpleExpression `json:"task"`
}

func New() *Application {
	logger.Init()

	return &Application{config: ConfigFromEnv()}

}

// Запуск агента
func (a *Application) RunAgent() error {

	ctx := context.Background()
	conn, err := grpc.Dial("localhost:5000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error(fmt.Sprintf("could not connect to grpc server:%v", err))

	}
	defer conn.Close()
	grpcClient := pb.NewCalculatorServiceClient(conn)

	jobs := make(chan calculator.SimpleExpression, 100)
	results := make(chan calculator.SimpleExpression, 100)
	//Создание воркеров нужного количества
	for w := 1; w <= a.config.COMPUTING_POWER; w++ {
		go Worker(w, jobs, results)
	}
	logger.Info(fmt.Sprintf("Agent start with COMPUTING_POWER:%d ", a.config.COMPUTING_POWER))
	for {
		select {
		case e := <-results:

			request := pb.PushResultRequest{
				Id:     e.Id,
				Result: float32(e.Result),
				Error:  e.Error,
			}
			grpcClient.PushResult(ctx, &request)
		default:

			response, err := grpcClient.GetTask(ctx, nil)
			if response == nil {
				continue
			}
			task := calculator.SimpleExpression{
				Id:             response.Id,
				Arg1:           response.Arg1,
				Arg2:           response.Arg2,
				Operation:      response.Operation,
				Operation_time: int(response.OperationTime),
			}
			if err == nil {

				jobs <- task
			} else if err == calculator.ErrServerNotWork {
				logger.Error(err.Error())
				return err
			}

		}
	}
	defer close(results)
	defer close(jobs)
	return nil

}
func Worker(id int, jobs <-chan calculator.SimpleExpression, results chan<- calculator.SimpleExpression) {
	for task := range jobs {
		logger.Info(fmt.Sprintf("Worker %d start task %s", id, task.Id))
		time.Sleep(time.Millisecond * time.Duration(task.Operation_time))
		task.Calc()
		results <- task
		logger.Info(fmt.Sprintf("Worker %d finished task %s with result %f", id, task.Id, task.Result))

	}
}
