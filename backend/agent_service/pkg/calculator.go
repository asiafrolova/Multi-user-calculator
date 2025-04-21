package calculator

import (
	"strconv"
)

type SimpleExpression struct {
	Id             string  `json:"id"`
	Arg1           string  `json:"arg1"`
	Arg2           string  `json:"arg2"`
	Operation      string  `json:"operation"`
	Operation_time int     `json:"operation_time"`
	Result         float64 `json:"-"`
	Error          string  `json:"-"`
}

// Подсчет выражения
func (e *SimpleExpression) Calc() error {

	newArg1, newArg2, err := e.ParseArg()
	if err != nil {
		e.Error = ErrInvalidArg.Error()
		return ErrInvalidArg
	}
	switch e.Operation {
	case "-":
		e.Result = newArg1 - newArg2
		return nil
	case "+":
		e.Result = newArg1 + newArg2
		return nil
	case "*":
		e.Result = newArg1 * newArg2
		return nil
	case "/":
		if newArg2 == 0 {
			e.Error = ErrDivisionByZero.Error()
			return ErrDivisionByZero
		}
		e.Result = newArg1 / newArg2
		return nil
	default:
		e.Error = ErrInvalidArg.Error()
		return ErrInvalidArg

	}
	e.Error = ErrInvalidArg.Error()

	return ErrInvalidArg
}

// Парсинг строчных аргументов
func (e *SimpleExpression) ParseArg() (float64, float64, error) {
	newArg1, err := strconv.ParseFloat(e.Arg1, 64)
	if err != nil {
		return 0, 0, err
	}
	newArg2, err := strconv.ParseFloat(e.Arg2, 64)
	if err != nil {
		return 0, 0, err
	}
	return newArg1, newArg2, nil

}
