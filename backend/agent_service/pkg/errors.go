package calculator

import "errors"

var (
	ErrDivisionByZero   = errors.New("Division by zero")          //Деление на ноль
	ErrInvalidOperation = errors.New("Invalid operation")         //Неизвестная операция
	ErrInvalidArg       = errors.New("Invalid arguments")         //Аргументы не являются числами
	ErrNotExpression    = errors.New("Not expressions")           //Нет выражений для вычисления
	ErrServerNotWork    = errors.New("Orchestrator doesn't work") //Оркестратор выключен или не отвечает

)
