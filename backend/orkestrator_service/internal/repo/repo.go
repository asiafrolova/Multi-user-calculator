package repo

import (
	"fmt"
	"strconv"

	"github.com/asiafrolova/Final_task/orkestrator_service/internal/database"
	orkestrator "github.com/asiafrolova/Final_task/orkestrator_service/pkg/orkestrator"
)

var (
	expressionsData   map[string]*orkestrator.Expression     //Хэш таблица с выражениями (ключ - id)
	lastID            int                                = 0 //Переменная для присвоения выражениям id
	SimpleExpressions chan orkestrator.SimpleExpression      //Канал для простых выражений
	currentExpression *orkestrator.Expression                //выражение с которым сейчас работаем
)

func Init() {
	//Инициализация осуществляемая только один раз
	if expressionsData == nil {
		expressionsData = make(map[string]*orkestrator.Expression)
	}
	if SimpleExpressions == nil {
		SimpleExpressions = make(chan orkestrator.SimpleExpression)
	}
	expressions, err := database.SelectExpressions()
	if err == nil {
		for _, elem := range expressions {
			expressionsData[elem.Id] = &elem
		}
	}
}

// Добавления выражения
func AddExpression(exp string, user_id int) (string, error) {
	var validExp bool = orkestrator.CheckExpression(exp)
	if !validExp {
		return "", orkestrator.ErrInvalidExpression
	}
	//Сохранение в базу данных
	expression := &orkestrator.Expression{Id: "", Exp: exp, Status: orkestrator.TODO, UserID: user_id}
	id, _ := database.InsertExpression(expression)
	currentID := strconv.Itoa(int(id))
	expressionsData[currentID] = expression

	return currentID, nil

}

// Удаление всех выражений пользователя
func DeleteExpressionsList(user_id int) int {
	count := 0
	for _, v := range expressionsData {
		if v == nil {
			continue
		}
		if v.UserID == user_id {
			delete(expressionsData, v.Id)
			id_int, err := strconv.Atoi(v.Id)
			if err != nil {
				continue
			}
			database.DeleteExpresssion(int64(id_int))
			count += 1
		}
	}
	return count

}

// Удаление выражения по id
func DeleteExpressionByID(id string, user_id int) error {
	exp, ok := expressionsData[id]
	if ok && exp.UserID == user_id {
		id_int, err := strconv.Atoi(id)
		if err != nil {
			return err
		}
		database.DeleteExpresssion(int64(id_int))
		delete(expressionsData, id)
		return nil
	}
	return orkestrator.ErrKeyExists
}

// Возврат выражения по id
func GetExpressionByID(id string, user_id int) (*orkestrator.Expression, error) {
	exp, ok := expressionsData[id]

	if ok && exp != nil && exp.UserID == user_id {
		return exp, nil
	}
	return &orkestrator.Expression{}, orkestrator.ErrKeyExists
}

// Возврат всех выражений
func GetExpressionsList(user_id int) []orkestrator.Expression {
	data := make([]orkestrator.Expression, 0)
	for _, v := range expressionsData {
		if v.UserID == user_id {
			data = append(data, *v)
		}
	}
	return data

}

// просим положить еще простых выражений канал
func GetSimpleOperations() {
	if currentExpression == nil || currentExpression.Status != orkestrator.PENDING {
		err := SetCurrentExpression()
		if err != nil {
			return
		}
	}

	for ind, elem := range currentExpression.SimpleExpressions {
		//Это выражение уже было отправлено агенту
		if elem.Processed {
			continue
		}

		newSimpleExpression, err := currentExpression.ConvertExpression(elem.Id)
		//Для выражения посчитаны нужные премененные
		if err == nil {

			currentExpression.SimpleExpressions[ind].Processed = true
			go func() {
				SimpleExpressions <- newSimpleExpression
			}()
		}
	}
}

// Устанавливаем выражение с которым будем работать
func SetCurrentExpression() error {
	for ind, elem := range expressionsData {

		if elem.Status == orkestrator.TODO {
			expressionsData[ind].Status = orkestrator.PENDING

			currentExpression = expressionsData[ind]
			database.UpdateExpression(currentExpression)
			tokenizeString, err := currentExpression.TokenizeString()
			if err != nil {

				currentExpression.Status = orkestrator.FAILED

				database.UpdateExpression(currentExpression)
				continue
			} else if len(tokenizeString) == 1 {
				currentExpression.Status = orkestrator.COMPLETED

				database.UpdateExpression(currentExpression)
				result, err := strconv.ParseFloat(tokenizeString[0], 64)
				if err != nil {
					currentExpression.Status = orkestrator.FAILED
					database.UpdateExpression(currentExpression)
				} else {
					currentExpression.Result = result
					database.UpdateExpression(currentExpression)
				}
				continue
			}
			_, err, _ = currentExpression.SplitExpression(tokenizeString)
			if err != nil {
				currentExpression.Status = orkestrator.FAILED
				database.UpdateExpression(currentExpression)
				continue
			}
			expressionsData[ind].WaitResult()
			return nil
		}
	}

	return orkestrator.ErrNotExpression
}

// Устанавливаем результат вычисления простого выражения и обновляем статус родительского выражения
func SetResult(id string, result float64, err error) error {

	if currentExpression == nil {
		return orkestrator.ErrNotExpression
	}
	if err != nil {
		currentExpression.Status = orkestrator.FAILED
		database.UpdateExpression(currentExpression)
		SetCurrentExpression()
		return nil
	}
	err = currentExpression.SetResultSimpleExpression(id, result)
	if err != nil {
		return err
	}
	//Для выражения посчитаны все подзадачи
	if len(currentExpression.SimpleExpressions) == len(currentExpression.SimpleExpressionsResults) {

		currentExpression.Status = orkestrator.COMPLETED
		//Ответ лежит в последней подзадаче (последнем действии)
		currentExpression.Result = currentExpression.SimpleExpressionsResults[currentExpression.SimpleExpressions[len(currentExpression.SimpleExpressions)-1].Id]
		database.UpdateExpression(currentExpression)
		SetCurrentExpression()
	}
	return nil

}

// Генерация id
func GenerateID() string {
	lastID++
	return fmt.Sprintf("id%d", lastID)
}
