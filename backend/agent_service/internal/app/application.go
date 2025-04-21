package application

import (
	"bytes"
	"encoding/json"

	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	logger "github.com/asiafrolova/Final_task/agent_service/internal/logger"
	calculator "github.com/asiafrolova/Final_task/agent_service/pkg"
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
			a.PushResult(&e)
		default:
			task, err := a.GetTask()
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

// Получение задачи
func (a *Application) GetTask() (calculator.SimpleExpression, error) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/internal/task", a.config.Addr))
	if resp == nil {
		return calculator.SimpleExpression{}, calculator.ErrServerNotWork
	}
	if resp.StatusCode == http.StatusOK {
		response := ResponseTask{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err == nil {
			logger.Info(fmt.Sprintf("task %v received", response.Task))
			return response.Task, nil
		} else {
			logger.Error(fmt.Sprintf("Unexpected error %v", err))
			return calculator.SimpleExpression{}, err
		}
	} else {

		return calculator.SimpleExpression{}, calculator.ErrNotExpression
	}
}

// Отправка ответа
func (a *Application) PushResult(e *calculator.SimpleExpression) {
	request := RequestResult{Id: e.Id, Result: e.Result, Err: e.Error}
	body, err := json.Marshal(request)
	if err != nil {
		logger.Error(fmt.Sprintf("Unexpected error %v", err))
	}
	resp, err := http.Post(fmt.Sprintf("http://localhost:%s/internal/task", a.config.Addr), "application/json", bytes.NewReader(body))
	if err != nil || resp.StatusCode != http.StatusOK {
		logger.Error(fmt.Sprintf("Unexpected error %v or bad StatusCode %v ", err, resp.StatusCode))

	}

}
