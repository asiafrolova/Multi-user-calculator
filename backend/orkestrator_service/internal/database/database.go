package database

import (
	"context"
	"database/sql"
	"strconv"
	"unicode"

	"github.com/asiafrolova/Final_task/orkestrator_service/internal/logger"
	"github.com/asiafrolova/Final_task/orkestrator_service/pkg/orkestrator"
	_ "github.com/mattn/go-sqlite3"
)

// глобальные переменные для работы с бд
var (
	ctx context.Context
	db  *sql.DB
)

// Структура для работы с пользоваталем
type User struct {
	Id       int
	Name     string
	Password string
}

// Инициализация
func Init() {
	logger.Init()
	var err error
	ctx = context.Background()

	db, err = sql.Open("sqlite3", "./data.db")

	if err != nil {

		logger.Error(err.Error())
		panic(err)
	}
	//defer db.Close()
	err = db.PingContext(ctx)

	if err != nil {

		logger.Error(err.Error())
		panic(err)
	}

	if err = createTables(); err != nil {

		logger.Error(err.Error())
		panic(err)
	}

}

// Создание таблиц для выражений и пользователей
func createTables() error {
	const (
		usersTable = `
	CREATE TABLE IF NOT EXISTS users(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE,
		password TEXT
	);`

		expressionsTable = `
	CREATE TABLE IF NOT EXISTS expressions(
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		expression TEXT NOT NULL,
		result INTEGER,
		user_id TEXT NOT NULL,
		status TEXT,


		FOREIGN KEY (user_id)  REFERENCES expressions (id)
	);`
	)

	if _, err := db.ExecContext(ctx, usersTable); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, expressionsTable); err != nil {
		return err
	}

	return nil
}

// Функции для вставки новых пользоваталей и выражений
func InsertUser(user *User) (error, int) {
	var q = `
	INSERT INTO users (name, password) values ($1, $2)
	`
	goodPassword := CheckPassword(user.Password)
	if !goodPassword {
		return orkestrator.ErrBadPassword, -1
	}
	if len(user.Name) < 4 {
		return orkestrator.ErrBadName, -1
	}
	result, err := db.ExecContext(ctx, q, user.Name, user.Password)
	if err != nil {
		return err, -1
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err, -1
	}
	user.Id = int(id)

	return nil, int(id)
}

// Функция для проверки пароля на надежность
func CheckPassword(password string) bool {
	digits := false
	letter := false
	for _, elem := range password {
		if unicode.IsDigit(elem) {
			digits = true
		}
		if unicode.IsLetter(elem) {
			letter = true
		}

	}
	if len(password) >= 8 && letter && digits {
		return true
	}
	return false
}
func InsertExpression(expression *orkestrator.Expression) (int64, error) {
	var q = `
	INSERT INTO expressions (expression, status, result, user_id) values ($1, $2,$3,$4)
	`
	result, err := db.ExecContext(ctx, q, expression.Exp, expression.Status, expression.Result, expression.UserID)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	expression.Id = strconv.FormatInt(id, 10)

	return id, nil
}

// Функции для получения списка всех пользоваталей и всех выражений
func SelectUsers() ([]User, error) {
	var users []User
	var q = "SELECT name, password FROM users"
	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		u := User{}
		err := rows.Scan(&u.Name, &u.Password)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

func SelectExpressions() ([]orkestrator.Expression, error) {
	var expressions []orkestrator.Expression
	var q = "SELECT id, expression,status,result, user_id FROM expressions"

	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		e := orkestrator.Expression{}
		err := rows.Scan(&e.Id, &e.Exp, &e.Status, &e.Result, &e.UserID)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, e)
	}

	return expressions, nil
}

// Функции для получения выражения или пользователя по id
func SelectUserByName(user_name string) (User, error) {
	u := User{}
	var q = "SELECT id,name, password FROM users WHERE name = $1"
	err := db.QueryRowContext(ctx, q, user_name).Scan(&u.Id, &u.Name, &u.Password)
	if err != nil {
		return u, err
	}

	return u, nil
}
func SelectExpressionByID(id int64) (orkestrator.Expression, error) {
	e := orkestrator.Expression{}
	var q = "SELECT id, expression, status,result,user_id FROM expressions WHERE id = $1"
	err := db.QueryRowContext(ctx, q, id).Scan(&e.Id, &e.Exp, &e.Status, &e.Result, &e.UserID)
	if err != nil {
		return e, err
	}

	return e, nil
}

// Функции для обновления значений user или expression в бд
func UpdateUser(user *User, user_id int) error {
	var q = "UPDATE users SET password=$1, name=$2 WHERE id = $3"
	var q2 = "UPDATE users SET name=$1 WHERE id = $2"
	var q3 = "UPDATE users SET password=$1 WHERE id = $2"
	if CheckPassword(user.Password) && len(user.Name) >= 4 {
		_, err := db.ExecContext(ctx, q, user.Password, user.Name, user_id)
		if err != nil {
			return err
		}
		return nil
	} else if len(user.Name) >= 4 {
		_, err := db.ExecContext(ctx, q2, user.Name, user_id)
		if user.Password == "" {
			return err
		} else {
			return orkestrator.ErrBadPassword
		}
	} else if CheckPassword(user.Password) {
		_, err := db.ExecContext(ctx, q3, user.Password, user_id)
		if user.Name == "" {
			return err
		} else {
			return orkestrator.ErrBadName
		}
	} else {
		return orkestrator.ErrBadName
	}

}
func UpdateExpression(e *orkestrator.Expression) error {
	var q = "UPDATE expressions SET expression = $1,status=$2,result=$3 WHERE id = $4"
	_, err := db.ExecContext(ctx, q, e.Exp, e.Status, e.Result, e.Id)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	return nil
}

// Функции для удаления user и expression
func DeleteUser(user_id int) error {
	var q = "DELETE FROM users WHERE id = $1"
	_, err := db.ExecContext(ctx, q, user_id)
	return err

}
func DeleteExpresssion(id int64) error {
	var q = "DELETE FROM expressions WHERE id = $1"
	_, err := db.ExecContext(ctx, q, id)
	return err

}
