package orkestrator

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

var (
	lastID                  int = 0
	TIME_ADDITION_MS        int
	TIME_SUBTRACTION_MS     int
	TIME_MULTIPLICATIONS_MS int
	TIME_DIVISIONS_MS       int
)

// Структура выражения
type Expression struct {
	Id                       string             `json:"id"`
	Exp                      string             `json:"-"`
	Status                   string             `json:"status"`
	Result                   float64            `json:"result"`
	SimpleExpressions        []SimpleExpression `json:"-"`
	SimpleExpressionsResults map[string]float64 `json:"-"`
	Timeout                  int                `json:"-"`
	mux                      sync.Mutex         `json:"-"`
	UserID                   int                `json:"userID"`
}

// Статусы выражений
var (
	TODO      string = "Todo"      //выражение ожидает
	PENDING   string = "Pending"   //выражение выполняется
	FAILED    string = "Failed"    //произошла ошибка
	COMPLETED string = "Completed" //выражение успешно вычислено
	TIMEOUT   string = "Timeout"   //ошибка по таймауту
)

// Структура подвыражния состящего из двух переменных и операции
type SimpleExpression struct {
	Id_parent      string `json:"-"`
	Id             string `json:"id"`
	Arg1           string `json:"arg1"`
	Arg2           string `json:"arg2"`
	Operation      string `json:"operation"`
	Operation_time int    `json:"operation_time"`
	Processed      bool   `json:"-"`
}

// Инициализация
func InitOrkestrator() {
	var err error
	TIME_ADDITION_MS, err = strconv.Atoi(os.Getenv("TIME_ADDITION_MS"))
	if err != nil {
		TIME_ADDITION_MS = 100
	}
	TIME_SUBTRACTION_MS, err = strconv.Atoi(os.Getenv("TIME_SUBTRACTION_MS"))
	if err != nil {
		TIME_SUBTRACTION_MS = 100
	}
	TIME_MULTIPLICATIONS_MS, err = strconv.Atoi(os.Getenv("TIME_MULTIPLICATIONS_MS"))
	if err != nil {
		TIME_MULTIPLICATIONS_MS = 100
	}
	TIME_DIVISIONS_MS, err = strconv.Atoi(os.Getenv("TIME_DIVISIONS_MS"))
	if err != nil {
		TIME_DIVISIONS_MS = 100
	}
}

// Первичная проверка выражения на корректность
func CheckExpression(exp string) bool {

	if !(unicode.IsDigit(rune(exp[len(exp)-1])) || exp[len(exp)-1] == ')') || !(unicode.IsDigit(rune(exp[0])) || exp[0] == '-' || exp[0] == '(') || strings.Contains(exp, "**") || strings.Contains(exp, "--") || strings.Contains(exp, "++") || strings.Contains(exp, "//") {

		return false

	}
	var balance int = 0
	for ind, elem := range exp {
		if elem == '.' {
			if !(ind != 0 && unicode.IsDigit(rune(exp[ind-1]))) {
				return false
			}
		} else if !(unicode.IsDigit(elem) || elem == '(' || elem == ')' || elem == '+' || elem == '-' || elem == '*' || elem == '/') {
			return false
		}
		if elem == '(' {
			balance++
		} else if elem == ')' {
			balance--
		}
		if balance < 0 {
			return false
		}
	}
	return balance == 0

}

// Деление строки на отдельные токены
func (e *Expression) TokenizeString() ([]string, error) {

	//Первичная проверка перед токенизацией
	if !CheckExpression(e.Exp) {
		return nil, ErrInvalidExpression
	}
	e.Exp = strings.TrimSpace(e.Exp)
	var result []string = make([]string, 0)
	var currentNum string = ""
	for ind, elem := range e.Exp {

		if unicode.IsDigit(elem) {
			currentNum += string(elem) //Если число
		} else if elem == '-' && (ind == 0 || e.Exp[ind-1] == '(') {
			//Если минус относится с числу

			result = append(result, "-1")
			result = append(result, "*")
		} else if elem == '.' && ind != 0 && len(currentNum) != 0 && !strings.Contains(currentNum, ".") {
			//Точка у вещественного числа
			currentNum += string(elem)
		} else if elem == '+' || elem == '-' || elem == '*' || elem == '/' {
			//Знаки операций могут идти только после чисел или закрывающей скобки

			if len(currentNum) == 0 {
				if ind == 0 || e.Exp[ind-1] != ')' {
					return nil, ErrInvalidExpression
				}
				result = append(result, string(elem))
				continue
			}
			result = append(result, currentNum)
			result = append(result, string(elem))
			currentNum = ""

		} else if elem == ')' || elem == '(' {
			//Остаются только скобки
			if len(currentNum) != 0 {
				result = append(result, currentNum)
				currentNum = ""

			}
			result = append(result, string(elem))
		} else {
			return nil, ErrInvalidExpression
		}
	}
	if e.Exp[len(e.Exp)-1] != ')' && len(currentNum) == 0 {
		return nil, ErrInvalidExpression
	}
	if len(currentNum) != 0 {
		result = append(result, currentNum)
	}

	return result, nil
}

// Присвоение результата выражению по индексу
func (e *Expression) SetResultSimpleExpression(id string, result float64) error {
	for _, elem := range e.SimpleExpressions {
		if elem.Id == id {
			e.SimpleExpressionsResults[id] = result
			return nil
		}
	}
	return ErrKeyExists
}

// Если все необходимые значения посчитаны можем преобразовать индексы выражений в их результаты
func (e *Expression) ConvertExpression(id string) (SimpleExpression, error) {
	for _, elem := range e.SimpleExpressions {
		if elem.Id == id {
			var result SimpleExpression = SimpleExpression{Id: elem.Id, Operation: elem.Operation, Operation_time: elem.Operation_time, Processed: elem.Processed}
			if strings.Contains(elem.Arg1, "se") {
				arg1, ok := e.SimpleExpressionsResults[elem.Arg1]
				if !ok {
					return SimpleExpression{}, ErrNotResult
				}
				result.Arg1 = strconv.FormatFloat(arg1, 'f', 6, 64)
			} else {
				result.Arg1 = elem.Arg1
			}
			if strings.Contains(elem.Arg2, "se") {
				arg2, ok := e.SimpleExpressionsResults[elem.Arg2]
				if !ok {
					return SimpleExpression{}, ErrNotResult
				}
				result.Arg2 = strconv.FormatFloat(arg2, 'f', 6, 64)
			} else {
				result.Arg2 = elem.Arg2
			}

			return result, nil

		}
	}
	return SimpleExpression{}, ErrKeyExists
}
func (e *Expression) SplitExpression(tokenizeString []string) ([]SimpleExpression, error, string) {
	lastExpID := ""
	//Создаем массив простых выражений если он еще не создан
	if e.SimpleExpressions == nil {

		e.SimpleExpressions = make([]SimpleExpression, 0)

	}
	if e.SimpleExpressionsResults == nil {
		e.SimpleExpressionsResults = make(map[string]float64, 0)
	}
	//Высший приоритет у скобок
	ind, containBrackets := FindInStringArr(tokenizeString, "(")
	for containBrackets {
		//ind2, _ := FindInStringArr(tokenizeString[ind+1:], ")")
		//ind2 += ind
		ind2, _ := FindPairBrackets(tokenizeString, ind)

		_, _, id := e.SplitExpression(tokenizeString[ind+1 : ind2])
		//Заменяем выражение в скобках на id операции из которой возьмем результат
		endString := make([]string, 0)
		if ind2+1 < len(tokenizeString) {
			endString = tokenizeString[ind2+1:]
		}
		startString := make([]string, 0)
		if ind > 0 {
			startString = tokenizeString[:ind]
		}
		tokenizeString = append(startString, id)
		tokenizeString = append(tokenizeString, endString...)

		//Снова ищем скобки в новом выражении
		ind, containBrackets = FindInStringArr(tokenizeString, "(")
	}
	//Следующие по приоритету операции * и /
	ind, containMult := FindInStringArr(tokenizeString, "*")
	ind2, containDiv := FindInStringArr(tokenizeString, "/")
	for containDiv || containMult {
		//Выбираем умножение или деление в зависимости от того какое действие встречается раньше
		if (ind > ind2 && ind2 != -1) || (ind == -1) {
			tmp := ind2
			ind2 = ind
			ind = tmp
		}

		id := GetID()
		if tokenizeString[ind] == "*" {
			e.SimpleExpressions = append(e.SimpleExpressions, SimpleExpression{Id_parent: e.Id, Id: id, Arg1: tokenizeString[ind-1], Arg2: tokenizeString[ind+1], Operation: tokenizeString[ind], Operation_time: TIME_MULTIPLICATIONS_MS})

		} else {
			e.SimpleExpressions = append(e.SimpleExpressions, SimpleExpression{Id_parent: e.Id, Id: id, Arg1: tokenizeString[ind-1], Arg2: tokenizeString[ind+1], Operation: tokenizeString[ind], Operation_time: TIME_DIVISIONS_MS})
		}
		//Заменяем выражение на его id
		endString := make([]string, 0)
		if ind+2 < len(tokenizeString) {
			endString = tokenizeString[ind+2:]
		}
		startString := make([]string, 0)
		if ind > 1 {
			startString = tokenizeString[:ind-1]
		}
		tokenizeString = append(startString, id)
		tokenizeString = append(tokenizeString, endString...)
		lastExpID = id
		//Снова ищем * и /
		ind, containMult = FindInStringArr(tokenizeString, "*")
		ind2, containDiv = FindInStringArr(tokenizeString, "/")

	}

	//Низший приоритет у операций + и -
	ind, containAdd := FindInStringArr(tokenizeString, "+")
	ind2, containSub := FindInStringArr(tokenizeString, "-")

	for containAdd || containSub {
		//Выбираем сложение или вычитание в зависимости от того какое действие встречается раньше
		if (ind > ind2 && ind2 != -1) || ind == -1 {
			tmp := ind2
			ind2 = ind
			ind = tmp
		}

		id := GetID()

		if tokenizeString[ind] == "+" {

			e.SimpleExpressions = append(e.SimpleExpressions, SimpleExpression{Id_parent: e.Id, Id: id, Arg1: tokenizeString[ind-1], Arg2: tokenizeString[ind+1], Operation: tokenizeString[ind], Operation_time: TIME_ADDITION_MS})
		} else {
			e.SimpleExpressions = append(e.SimpleExpressions, SimpleExpression{Id_parent: e.Id, Id: id, Arg1: tokenizeString[ind-1], Arg2: tokenizeString[ind+1], Operation: tokenizeString[ind], Operation_time: TIME_SUBTRACTION_MS})
		}

		//Заменяем выражение на его id
		endString := make([]string, 0)
		if ind+2 < len(tokenizeString) {
			endString = tokenizeString[ind+2:]
		}
		startString := make([]string, 0)
		if ind > 1 {
			startString = tokenizeString[:ind-1]
		}
		tokenizeString = append(startString, id)
		tokenizeString = append(tokenizeString, endString...)
		lastExpID = id
		//Снова ищем - и +
		ind, containAdd = FindInStringArr(tokenizeString, "+")
		ind2, containSub = FindInStringArr(tokenizeString, "-")

	}

	//Время максимального вычисление выражения
	e.Timeout = max(TIME_ADDITION_MS, TIME_DIVISIONS_MS, TIME_MULTIPLICATIONS_MS, TIME_SUBTRACTION_MS) * (len(e.SimpleExpressions) * 2)
	//возвращаем последовательность выражений
	return e.SimpleExpressions, nil, lastExpID

}

// Поиск элемента в массиве строк
func FindInStringArr(input []string, item string) (int, bool) {
	for ind, elem := range input {
		if elem == item {
			return ind, true
		}

	}
	return -1, false
}

// Поиск закрывающей скобки для открывающей скобки на позиции indStartBrackets
func FindPairBrackets(input []string, indStartBrackets int) (int, bool) {
	balance := 0
	for ind := indStartBrackets; ind < len(input); ind++ {
		if input[ind] == "(" {
			balance++
		} else if input[ind] == ")" {
			balance--
		}
		if balance == 0 {
			return ind, true
		}
	}

	return -1, false

}

// Если при вычисление выражения будет начато, но агент не отправит ответ в течение таймаута, вычисление будет остановлено
func (e *Expression) WaitResult() {
	go func() {
		timer := time.NewTimer(time.Millisecond * time.Duration(e.Timeout))
		for {
			select {
			case <-timer.C:
				e.mux.Lock()
				e.Status = TIMEOUT
				e.mux.Unlock()
				return
			default:
				e.mux.Lock()
				if e.Status != PENDING {
					return
				}
				e.mux.Unlock()
			}
		}
	}()
}

func GetID() string {
	lastID++
	return fmt.Sprintf("se%d", lastID)

}
