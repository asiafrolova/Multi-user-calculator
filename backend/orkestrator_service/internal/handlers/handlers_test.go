package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/asiafrolova/Final_task/orkestrator_service/internal/database"
	"github.com/asiafrolova/Final_task/orkestrator_service/internal/handlers"
	"github.com/asiafrolova/Final_task/orkestrator_service/internal/logger"
	"github.com/asiafrolova/Final_task/orkestrator_service/pkg/orkestrator"
)

func RegisterAndLoginUser() string {
	logger.Init()
	postBody := map[string]interface{}{
		"name":     "Vasya",
		"password": "qwerty",
	}
	body, _ := json.Marshal(postBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/register", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handlers.RegisterUser(w, req)
	postBody = map[string]interface{}{
		"name":     "Vasya",
		"password": "qwerty",
	}
	body, _ = json.Marshal(postBody)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader(body))
	w = httptest.NewRecorder()
	handlers.LoginUser(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	var respBody *map[string]string

	err = json.Unmarshal(body, &respBody)
	if err != nil {
		return ""
	}
	token := ""
	for key, val := range *respBody {
		if key == "access_token" {
			token = fmt.Sprint(val)
			break
		}
	}
	return token

}
func DeleteUser(token string) {
	bearer := "Bearer " + token
	req := httptest.NewRequest(http.MethodPost, "/api/v1/delete", nil)
	w := httptest.NewRecorder()
	req.Header.Set("Authorization", bearer)
	handlers.DeleteUser(w, req)
}

func TestRegisterUser(t *testing.T) {
	logger.Init()
	user1 := database.User{Name: "Vasya", Password: "qwerty1234"}
	postBody := map[string]interface{}{
		"name":     user1.Name,
		"password": user1.Password,
	}
	body, _ := json.Marshal(postBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/register", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handlers.RegisterUser(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected error statuscode: %v", resp.StatusCode)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	var respBody *map[string]interface{}

	err = json.Unmarshal(body, &respBody)
	if err != nil {
		t.Fatalf("Unexpected error  %v", err)
	}
	name := ""
	for key, val := range *respBody {
		if key == "name" {
			name = fmt.Sprint(val)
			break
		}
	}
	if name != user1.Name {
		t.Fatalf("Invalid name wan: %v, got: %v", user1.Name, name)
	}
	postBody = map[string]interface{}{
		"name":     user1.Name,
		"password": user1.Password,
	}
	body, _ = json.Marshal(postBody)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader(body))
	w = httptest.NewRecorder()
	handlers.LoginUser(w, req)
	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected error statuscode: %v", resp.StatusCode)
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	err = json.Unmarshal(body, &respBody)
	if err != nil {
		t.Fatalf("Unexpected error  %v", err)
	}
	token := ""
	for key, val := range *respBody {
		if key == "access_token" {
			token = fmt.Sprint(val)
			break
		}
	}

	bearer := "Bearer " + token
	req = httptest.NewRequest(http.MethodPost, "/api/v1/delete", bytes.NewReader(body))
	w = httptest.NewRecorder()
	req.Header.Set("Authorization", bearer)
	handlers.DeleteUser(w, req)
	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected error statuscode: %v", resp.StatusCode)
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	err = json.Unmarshal(body, &respBody)
	if err != nil {
		t.Fatalf("Unexpected error  %v", err)
	}
	name = ""
	for key, val := range *respBody {
		if key == "name" {
			name = fmt.Sprint(val)
			break
		}
	}

}

func TestAddExpressionGetIDHandlerOK(t *testing.T) {
	token := RegisterAndLoginUser()
	bearer := "Bearer " + token

	testCases := []struct {
		input string
	}{
		{"1+2"},
		{"1+2+3"},
		{"((1+1)-(2-1))"},
		{"3.14-1"},
		{"1/1"},
	}
	for _, tc := range testCases {
		logger.Init()
		postBody := map[string]interface{}{
			"expression": tc.input,
		}
		body, _ := json.Marshal(postBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewReader(body))
		w := httptest.NewRecorder()

		req.Header.Set("Authorization", bearer)

		handlers.AddExpressionsHandler(w, req)
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Unexpected error with expression %v, statuscode: %v", tc.input, resp.StatusCode)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Unexpected error with expression %v: %v", tc.input, err)
		}
		var respBody *map[string]string

		err = json.Unmarshal(body, &respBody)
		if err != nil {
			t.Fatalf("Unexpected error with expression %v : %v", tc.input, err)
		}
		id := ""
		for key, val := range *respBody {
			if key == "id" {
				id = fmt.Sprint(val)
				break
			}
		}
		if id == "" {
			t.Fatalf("Invalid id with expression %v", tc.input)
		}

		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/expressions?id=%s", id), nil)
		w = httptest.NewRecorder()
		req.Header.Set("Authorization", bearer)
		handlers.GetExpressionByIDHandler(w, req)
		resp = w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Unexpected error with expression %v, statuscode: %v", tc.input, resp.StatusCode)
		}

	}
	DeleteUser(token)

}
func TestAddExpressionGetIDHandlerError(t *testing.T) {
	token := RegisterAndLoginUser()
	bearer := "Bearer " + token
	testCases := []struct {
		input string
		want  error
	}{
		{"1+2)", orkestrator.ErrInvalidExpression},
		{"1+2**3", orkestrator.ErrInvalidExpression},
		{"((1+101))))-(2-1))", orkestrator.ErrInvalidExpression},
		{"3.14-1-~3", orkestrator.ErrInvalidExpression},
		{"1/1a", orkestrator.ErrInvalidExpression},
	}
	for _, tc := range testCases {
		logger.Init()
		postBody := map[string]interface{}{
			"expression": tc.input,
		}
		body, _ := json.Marshal(postBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewReader(body))
		w := httptest.NewRecorder()
		req.Header.Set("Authorization", bearer)
		handlers.AddExpressionsHandler(w, req)
		resp := w.Result()
		if resp.StatusCode == http.StatusOK {
			t.Fatalf("Invalid status with expression %v, statuscode: %v", tc.input, resp.StatusCode)
		}

		id := ""

		req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/expressions?id=%s", id), nil)
		w = httptest.NewRecorder()
		req.Header.Set("Authorization", bearer)
		handlers.GetExpressionByIDHandler(w, req)
		resp = w.Result()
		if resp.StatusCode == http.StatusOK {
			t.Fatalf("Invalid status with expression %v, statuscode: %v", tc.input, resp.StatusCode)
		}

	}
	DeleteUser(token)

}
