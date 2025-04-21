package database_test

import (
	"testing"

	"github.com/asiafrolova/Multi-user-calculator/orkestrator_service/internal/database"
	"github.com/asiafrolova/Multi-user-calculator/orkestrator_service/pkg/orkestrator"
)

func TestUserTable(t *testing.T) {
	database.Init()
	user := database.User{Name: "Vasya", Password: "qwerty1234"}
	err, id := database.InsertUser(&user)

	if err != nil {
		t.Errorf(err.Error())
	}
	get_user, err := database.SelectUserByName(user.Name)
	if err != nil {
		t.Errorf(err.Error())
	}
	if get_user.Name != user.Name || get_user.Password != user.Password {
		t.Errorf("Invalid data, exprected: %v, got: %v", user, get_user)
	}

	user.Password = "1234qwerty"
	err = database.UpdateUser(&user, id)
	if err != nil {
		t.Errorf(err.Error())
	}
	get_user, err = database.SelectUserByName(user.Name)

	if get_user.Name != user.Name || get_user.Password != user.Password {
		t.Errorf("Invalid data, exprected: %v, got: %v", user, get_user)
	}
	two_user := database.User{Name: "Petr", Password: "qwerty1234"}
	err, id2 := database.InsertUser(&two_user)

	if err != nil {
		t.Errorf(err.Error())
	}
	all_users, err := database.SelectUsers()
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(all_users) != 2 {
		t.Errorf("Bad size data users, exptected %v, got %v", 2, len(all_users))
	}
	err = database.DeleteUser(id)
	if err != nil {
		t.Errorf(err.Error())
	}
	get_user, err = database.SelectUserByName(user.Name)
	if err == nil {
		t.Errorf("Error delete user")
	}
	database.DeleteUser(id2)

}
func TestExpressionsTable(t *testing.T) {
	database.Init()
	expression := orkestrator.Expression{Exp: "1+1", Status: orkestrator.TODO, UserID: 1}
	id, err := database.InsertExpression(&expression)
	if err != nil {
		t.Errorf(err.Error())
	}
	expression2 := orkestrator.Expression{Exp: "1*3", Status: orkestrator.TODO, UserID: 1}
	id2, err := database.InsertExpression(&expression2)
	if err != nil {
		t.Errorf(err.Error())
	}
	get_exp1, err := database.SelectExpressionByID(id)
	if err != nil {
		t.Errorf(err.Error())
	}
	if get_exp1.Id != expression.Id || get_exp1.Exp != expression.Exp || get_exp1.Status != expression.Status || get_exp1.UserID != expression.UserID {
		t.Errorf("Error SelectExpression(), want %v, got %v", expression, get_exp1)
	}
	all_exp, err := database.SelectExpressions()
	if len(all_exp) != 2 {
		t.Errorf("Bad size data users, exptected %v, got %v", 2, len(all_exp))
	}
	database.DeleteExpresssion(id)
	database.DeleteExpresssion(id2)
	all_exp, err = database.SelectExpressions()
	if len(all_exp) != 0 {
		t.Errorf("Bad size data users, exptected %v, got %v", 0, len(all_exp))
	}

}
