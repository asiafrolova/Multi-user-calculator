package orkestrator

import "errors"

var (
	ErrInvalidExpression = errors.New("Invaild expression")                     //Ошибка в выражении
	ErrKeyExists         = errors.New("Invalid id")                             //Выражения с таким id не существует
	ErrNotResult         = errors.New("Expression has not yet been calculated") //Выражение ещё не посчитано
	ErrNotExpression     = errors.New("Not expressions")                        //Нет выражений для вычисления
	ErrBadCredentials    = errors.New("Name or password is incorrect")          //Неверный пароль или имя
	ErrBadPassword       = errors.New("Bad password")                           //Плохой пароль, слишком короткий или ненадежный
	ErrBadName           = errors.New("Bad name")
)
