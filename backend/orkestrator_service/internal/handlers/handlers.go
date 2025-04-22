package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/asiafrolova/Multi-user-calculator/orkestrator_service/internal/database"
	logger "github.com/asiafrolova/Multi-user-calculator/orkestrator_service/internal/logger"
	repo "github.com/asiafrolova/Multi-user-calculator/orkestrator_service/internal/repo"
	orkestrator "github.com/asiafrolova/Multi-user-calculator/orkestrator_service/pkg/orkestrator"
	jwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	WAITING_TIME = time.Millisecond * 100    //Ожидание появление новой задачи для отправки агенту
	jwtSecretKey = []byte("very-secret-key") // Секретный ключ для подписи JWT-токена
)

// Запрос на добавление выражения
type RequestAddExpression struct {
	Expression string `json:"expression"`
}

// Ответ после успешного добавление выражения
type ResponseAddExpression struct {
	Id string `json:"id"`
}

// Ответ на запрос листа выражений
type ResponseListExpressions struct {
	Expressions []orkestrator.Expression `json:"expressions"`
}

// Ответ на запрос выражения по id
type ResponseIDExpression struct {
	Expression orkestrator.Expression `json:"expression"`
}

// Ответ на запрос задачи (простого выражения с одним действием)
type ResponseGetTask struct {
	Task orkestrator.SimpleExpression `json:"task"`
}

// Структура HTTP-запроса на регистрацию пользователя
type RegisterRequest struct {
	Name     string `json:"login"`
	Password string `json:"password"`
}
type RegisterResponse struct {
	Name string `json:"login"`
	Id   int    `json:"id"`
}

// Структура HTTP-запроса на вход в аккаунт
type LoginRequest struct {
	Name     string `json:"login"`
	Password string `json:"password"`
}

// Структура HTTP-ответа на вход в аккаунт
// В ответе содержится JWT-токен авторизованного пользователя
type LoginResponse struct {
	AccessToken string `json:"access_token"`
	LastTime    int64  `json:"time,string"`
}

// Структура HTTP-ответа на удаления аккаунта
type DeleteResponse struct {
	Id int `json:"id"`
}

// Запрос на добавление результата простого выражения
type RequestAddResult struct {
	Id     string  `json:"id"`
	Result float64 `json:"result"`
	Err    string  `json:"error"`
}

// Структура HTTP-ответа с информацией о пользователе
type ProfileResponse struct {
	Password string `json:"password"`
	Name     string `json:"login"`
}

// Структура ответа с информацией о удалении
type DeleteExpressionsResponse struct {
	Count int    `json:"count"`
	Err   string `json:"error"`
}

func WithCORS(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Проверка токена
func CheckToken(authHeader string) (int, error) {

	if authHeader == " " {
		logger.Error(orkestrator.ErrBadCredentials.Error())
		return -1, orkestrator.ErrBadCredentials
	}
	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		logger.Error(orkestrator.ErrBadCredentials.Error())
		return -1, orkestrator.ErrBadCredentials
	}
	tokenFromString, err := jwt.Parse(headerParts[1], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			panic(fmt.Errorf("unexpected signing method: %v", token.Header["alg"]))
		}

		return []byte(jwtSecretKey), nil
	})
	if err != nil {
		logger.Error(orkestrator.ErrBadCredentials.Error())
		return -1, orkestrator.ErrBadCredentials
	}

	claims, ok := tokenFromString.Claims.(jwt.MapClaims)
	if !ok {
		logger.Error(orkestrator.ErrBadCredentials.Error())
		return -1, orkestrator.ErrBadCredentials
	}

	logger.Info(fmt.Sprintf("Token correct, name: %v", claims["id"]))
	if err != nil {
		return 0, err
	}

	id, err := strconv.Atoi(fmt.Sprint(claims["id"]))

	return id, nil
}

// хэндлер для удаления всех выражений пользователя
func DeleteExpressions(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has("id") {
		DeleteExpressionByID(w, r)
		return
	}
	//проверка прав пользователя
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	user_id, err := CheckToken(r.Header.Get("Authorization"))
	if err != nil {
		logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusOK)
		return
	}

	count := repo.DeleteExpressionsList(user_id)
	response := DeleteExpressionsResponse{Count: count, Err: ""}
	res, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) //Ошибка при сериализации
	}
	w.WriteHeader(http.StatusOK)
	w.Write(res)
	logger.Info(fmt.Sprintf("Expressions by user with id %v deleted successfully", user_id))

}

func DeleteExpressionByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	//проверка прав пользователя
	user_id, err := CheckToken(r.Header.Get("Authorization"))

	if err != nil {
		logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusOK)
		return
	}

	id := r.PathValue("id")
	if id == "" && r.URL.Query().Has("id") {
		id = r.URL.Query().Get("id")

	}
	err = repo.DeleteExpressionByID(id, user_id)
	response := DeleteExpressionsResponse{Count: 1, Err: ""}
	if err != nil {
		response = DeleteExpressionsResponse{Count: 0, Err: err.Error()}
	}

	res, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) //Ошибка при сериализации
	}
	w.WriteHeader(http.StatusOK)
	w.Write(res)
	logger.Info(fmt.Sprintf("Expressions by user with id %v with id %v deleted successfully", user_id, id))

}

// Хэндлер для обновления данных пользователя
func UpdateUser(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	user_id, err := CheckToken(r.Header.Get("Authorization"))

	if err != nil {
		logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusOK)
		return
	}
	request := new(RegisterRequest)
	defer r.Body.Close()

	err = json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError) //Ошибка при десериализации
		return
	}

	gen_password, err := generate(request.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) //Ошибка при шифровании пароля
		return
	}
	user := database.User{Name: request.Name, Password: gen_password}
	if request.Password == "" {
		user = database.User{Name: request.Name, Password: ""}
	}
	database.Init()
	err = database.UpdateUser(&user, user_id)
	if err != nil {
		logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest) //Ошибка при регистрации
		return

	}
	payload := jwt.MapClaims{
		"name": user.Name,
		"exp":  time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	t, err := token.SignedString(jwtSecretKey)
	if err != nil {
		logger.Error("JWT token signing")
		http.Error(w, orkestrator.ErrBadCredentials.Error(), http.StatusInternalServerError)
		return
	}
	response := LoginResponse{AccessToken: t}
	w.WriteHeader(http.StatusOK)
	res, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) //Ошибка при сериализации
		return
	}
	_, err = w.Write(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info(fmt.Sprintf("User with id %v updated, new name - %v", user_id, user.Name))
}

// Хэндлер для удаления пользователя
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	user_id, err := CheckToken(r.Header.Get("Authorization"))
	if err != nil {
		logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusOK)
		return
	}
	repo.DeleteExpressionsList(user_id)
	err = database.DeleteUser(user_id)
	if err != nil {
		logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := DeleteResponse{Id: user_id}
	w.WriteHeader(http.StatusOK)
	res, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) //Ошибка при сериализации
		return
	}
	_, err = w.Write(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info(fmt.Sprintf("User with id %v deleted", response.Id))

}

// Хэндлер для входа пользователей
func LoginUser(w http.ResponseWriter, r *http.Request) {
	request := LoginRequest{}
	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError) //Ошибка при десериализации
		return
	}
	user, err := database.SelectUserByName(request.Name)
	if err != nil {
		logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest) //Ошибка авторизации
		return
	}
	err = compare(user.Password, request.Password)
	if err != nil {
		logger.Error(orkestrator.ErrBadCredentials.Error() + "1")
		http.Error(w, orkestrator.ErrBadCredentials.Error(), http.StatusBadRequest) //Неверный пароль
		return
	}
	// Генерируем JWT-токен для пользователя,
	// который он будет использовать в будущих HTTP-запросах

	payload := jwt.MapClaims{
		"id":  user.Id,
		"exp": time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	t, err := token.SignedString(jwtSecretKey)
	if err != nil {
		logger.Error("JWT token signing")
		http.Error(w, orkestrator.ErrBadCredentials.Error(), http.StatusInternalServerError)
		return
	}
	response := LoginResponse{AccessToken: t, LastTime: time.Now().Add(time.Hour * 72).Unix()}
	w.WriteHeader(http.StatusOK)
	res, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) //Ошибка при сериализации
		return
	}
	_, err = w.Write(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info(fmt.Sprintf("User with name %v signing", user.Name))
}

// Хэндлер для регистрации пользователей
func RegisterUser(w http.ResponseWriter, r *http.Request) {

	request := new(RegisterRequest)
	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError) //Ошибка при десериализации
		return
	}
	gen_password, err := generate(request.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) //Ошибка при шифровании пароля
		return
	}
	user := database.User{Name: request.Name, Password: gen_password}
	database.Init()
	err, id := database.InsertUser(&user)
	if err != nil {
		logger.Error(fmt.Sprintf("%v |%v|", err.Error(), request))
		http.Error(w, err.Error(), http.StatusBadRequest) //Ошибка при регистрации
		return

	}
	response := RegisterResponse{Name: user.Name, Id: id}
	w.WriteHeader(http.StatusOK)
	res, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) //Ошибка при сериализации
		return
	}
	_, err = w.Write(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info(fmt.Sprintf("User with name %v registred", response.Name))

}

// Хендлер для добавления выражения
func AddExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	//проверка прав пользователя
	user_id, err := CheckToken(r.Header.Get("Authorization"))
	if err != nil {
		logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusOK)
		return
	}

	request := new(RequestAddExpression)
	defer r.Body.Close()

	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError) //Ошибка при десериализации
		return
	}
	repo.Init()

	id, err := repo.AddExpression(request.Expression, user_id)
	if err != nil {
		if err == orkestrator.ErrInvalidExpression {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity) //Выражение не прошло первичную проверку
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError) //Неизвестная ошибка
			return
		}
	}

	response := ResponseAddExpression{Id: id}
	w.WriteHeader(http.StatusOK)
	res, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) //Ошибка при сериализации
		return
	}
	_, err = w.Write(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info(fmt.Sprintf("Expression with id %s added", response.Id))

}

// Хендлер для получения списка выражений
func GetExpressionsListHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	if r.URL.Query().Has("id") {
		GetExpressionByIDHandler(w, r)
		return
	}
	//проверка прав пользователя

	user_id, err := CheckToken(r.Header.Get("authorization"))
	if err != nil {
		logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusOK)
		return
	}

	repo.Init()
	response := ResponseListExpressions{Expressions: repo.GetExpressionsList(user_id)}
	res, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) //Ошибка при сериализации
	}
	w.WriteHeader(http.StatusOK)
	w.Write(res)
	logger.Info("List of expressions returned successfully")

}

// Хендлер для получения выражения по id
func GetExpressionByIDHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	//проверка прав пользователя
	user_id, err := CheckToken(r.Header.Get("Authorization"))
	if err != nil {
		logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	w.Header().Set("Access-Control-Allow-Origin", "*")

	repo.Init()
	id := r.PathValue("id")
	if id == "" && r.URL.Query().Has("id") {
		id = r.URL.Query().Get("id")

	}
	exp, err := repo.GetExpressionByID(id, user_id)

	if err == orkestrator.ErrKeyExists {
		http.Error(w, err.Error(), http.StatusNotFound) //Нет такого ключа
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) //Неизвестная ошибка
		return
	} else if exp == nil {
		http.Error(w, orkestrator.ErrNotExpression.Error(), http.StatusNotFound) //Выражение удалено
		return
	}

	response := ResponseIDExpression{}
	response.Expression = *exp

	res, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) //Ошибка при сериализации
	}
	w.WriteHeader(http.StatusOK)
	w.Write(res)
	logger.Info("Expression by id returned successfully")
}

func generate(s string) (string, error) {
	saltedBytes := []byte(s)
	hashedBytes, err := bcrypt.GenerateFromPassword(saltedBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	hash := string(hashedBytes[:])
	return hash, nil
}
func compare(hash string, s string) error {
	incoming := []byte(s)
	existing := []byte(hash)
	return bcrypt.CompareHashAndPassword(existing, incoming)
}
