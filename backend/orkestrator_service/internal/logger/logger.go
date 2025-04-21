package logger

import (
	"log/slog"
	"os"
)

type logger struct {
	log *slog.Logger
}

var (
	globalLogger logger
)

func Init() {
	//Инициализация происходит только один раз
	if globalLogger.log == nil {

		log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		globalLogger.log = log
	}

}

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

func Debug(msg string) {
	globalLogger.log.Debug(msg)
}

func Info(msg string) {
	globalLogger.log.Info(msg)
}

func Warn(msg string) {
	globalLogger.log.Warn(msg)
}

func Error(msg string) {
	globalLogger.log.Error(msg)
}
