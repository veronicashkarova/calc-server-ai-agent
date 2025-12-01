package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/veronicashkarova/server-for-calc/pkg/calc"
	"github.com/veronicashkarova/server-for-calc/pkg/contract"
	"github.com/veronicashkarova/server-for-calc/pkg/db"
	"github.com/veronicashkarova/server-for-calc/pkg/orkestrator"
)

func RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	request := new(contract.UserLogin)
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newUser := contract.UserLogin{
		Login:    request.Login,
		Password: request.Password,
	}

	err = orkestrator.RegisterUser(&newUser)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	request := new(contract.UserLogin)
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := contract.UserLogin{
		Login:    request.Login,
		Password: request.Password,
	}

	token, tokenErr := orkestrator.LoginUser(&user)

	if tokenErr != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, token)
}

func AutorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		userLogin, err := orkestrator.CheckToken(authHeader)

		if err != nil {
			http.Error(w, errors.New("FAILED AUTORIZATION").Error(), http.StatusBadRequest)
		} else {
			ctx := context.WithValue(r.Context(), "user_login", userLogin)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func NewExpressionHandler(w http.ResponseWriter, r *http.Request) {
	request := new(Request)
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userLogin := r.Context().Value("user_login").(string)
	result, id, err := orkestrator.AddExpression(userLogin, request.Expression)

	if err != nil {
		switch {
		case errors.Is(err, calc.ErrInvalidExpression):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, calc.ErrEmptyExpression):
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
	} else {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, result)

		go func() {
			fmt.Printf("NewExpressionHandler: запуск calc.Calc для выражения %s с ID=%s\n", request.Expression, id)
			result, err := calc.Calc(request.Expression, id, contract.TaskChannel)
			fmt.Printf("NewExpressionHandler: calc.Calc завершился для ID=%s, result=%f, err=%v\n", id, result, err)
			value, exist := contract.ExpressionMap[id]
			if exist {
				if err != nil {
					fmt.Printf("NewExpressionHandler: ошибка вычисления для ID=%s: %v\n", id, err)
					value.Data.Status = err.Error()
				} else {
					fmt.Printf("NewExpressionHandler: вычисление успешно для ID=%s, результат=%f\n", id, result)
					value.Data.Status = contract.Done
					value.Data.Result = strconv.FormatFloat(result, 'f', 3, 64)
				}
				contract.ExpressionMap[id] = value

				intId, err := strconv.ParseInt(id, 10, 64)
				if err == nil {
					db.UpdateExpressionStatusResult(intId, contract.Done, strconv.FormatFloat(result, 'f', 3, 64))
				}
			} else {
				fmt.Printf("NewExpressionHandler: выражение ID=%s не найдено в ExpressionMap\n", id)
			}
		}()
	}
}

func ExpressionsHandler(w http.ResponseWriter, r *http.Request) {

	userLogin := r.Context().Value("user_login").(string)
	result, err := orkestrator.Expressions(userLogin)
	if err != nil {
		switch {
		case errors.Is(err, calc.ErrInvalidExpression):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, calc.ErrEmptyExpression):
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, result)
	}
}

func IdHandler(w http.ResponseWriter, r *http.Request) {
	id, err := isIdExpressionRequest(r.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	userLogin := r.Context().Value("user_login").(string)
	result, err := orkestrator.GetExpressionForId(userLogin, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, result)
}

func isIdExpressionRequest(url *url.URL) (string, error) {

	// Разделяем путь на сегменты
	pathSegments := strings.Split(url.Path, "/")

	// Получаем последний сегмент
	lastSegment := pathSegments[len(pathSegments)-1]
	fmt.Println(lastSegment)

	return lastSegment, nil
}
