package orkestrator

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/veronicashkarova/server-for-calc/pkg/calc"
	"github.com/veronicashkarova/server-for-calc/pkg/contract"
	"github.com/veronicashkarova/server-for-calc/pkg/db"

	"github.com/golang-jwt/jwt/v5"
)

func RegisterUser(user *contract.UserLogin) error {
	_, err := db.InsertUser(user)
	return err
}

func LoginUser(user *contract.UserLogin) (string, error) {

	if !db.CheckUser(user) {
		return "", errors.New("USER NOT REGISTERED")
	}

	token, err := getToken(user.Login)

	if err != nil {
		return "tokenData", err
	}

	tokenData := contract.TokenData{Token: token}
	jsonBytes, err := json.Marshal(tokenData)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func AddExpression(userLogin string, expression string) (string, string, error) {
	var id int64
	userId, err := db.SelectIdForUser(userLogin)

	if err == nil {
		dbExpression := db.Expression{
			ID:         int64(id),
			Expression: expression,
			UserID:     userId,
			Status:     contract.InProcess,
			Result:     contract.Undefined,
		}
		id, err = db.InsertExpression(&dbExpression)
		if err != nil {
			fmt.Println(err)
		}
	}

	newId := strconv.Itoa(int(id))
	expressionData :=
		contract.ExpressionData{
			ID:     newId,
			Status: contract.InProcess,
			Result: contract.Undefined,
		}

	contract.ExpressionMap[newId] = contract.ExpressionMapData{
		User:    userLogin,
		Data:    expressionData,
		ExpChan: make(chan float64),
	}

	response := contract.ResponseData{ID: newId}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}

	return string(jsonBytes), newId, nil

}

func Expressions(userLogin string) (string, error) {
	var expressionsData []contract.ExpressionData

	userId, err := db.SelectIdForUser(userLogin)

	if err == nil {
		expressions, err := db.SelectExpressionsForUserId(userId)
		if err == nil {
			for _, expression := range expressions {
				expressionsData = append(
					expressionsData,
					contract.ExpressionData{
						ID:     fmt.Sprint(expression.ID),
						Status: expression.Status,
						Result: expression.Result,
					})
			}
		}
	}

	jsonBytes, err := json.Marshal(expressionsData)
	if err != nil {
		panic(err)
	}

	return string(jsonBytes), nil
}

func GetExpressionForId(userLogin string, id string) (string, error) {

	expression, error := findExpressionForId(userLogin, id)

	if error == nil {
		jsonBytes, err := json.Marshal(expression)
		if err != nil {
			panic(err)
		}
		return string(jsonBytes), nil
	}
	return "", error
}

func GetTaskData() (contract.TaskData, error) {
	select {
	case task := <-contract.TaskChannel:
		fmt.Printf("GetTaskData: получена задача из канала: ID=%d, Arg1=%f, Arg2=%f, Operation=%s\n", task.ID, task.Arg1, task.Arg2, task.Operation)
		return task, nil
	default:
		fmt.Printf("GetTaskData: канал пуст, задач нет\n")
		return contract.TaskData{}, calc.ErrNotTask
	}
}

func findExpressionForId(userLogin string, id string) (contract.ExpressionData, error) {
	var expressionData contract.ExpressionData
	userId, err := db.SelectIdForUser(userLogin)

	if err == nil {
		intId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return expressionData, err
		}
		expression, err := db.SelectExpressionForId(intId)
		if err != nil {
			return expressionData, err
		}

		if userId == expression.UserID {
			expressionData = contract.ExpressionData{
				ID:     fmt.Sprint(expression.ID),
				Status: expression.Status,
				Result: expression.Result,
			}
		}

		return expressionData, nil
	}

	name, found := contract.ExpressionMap[id]
	if !found {
		return contract.ExpressionData{}, calc.ErrNotFound
	}

	return name.Data, nil
}

func SendResult(id int, result float64) error {
	strId := fmt.Sprint(id)
	fmt.Printf("SendResult: получен результат для задачи ID=%d (выражение %s): %f\n", id, strId)
	_, exists := contract.ExpressionMap[strId]
	if exists {
		task := contract.ExpressionMap[strId].Data
		if task.Result != contract.Done {
			fmt.Printf("SendResult: отправка результата %f в ExpChan для выражения %s\n", result, strId)
			contract.ExpressionMap[strId].ExpChan <- result
			fmt.Printf("SendResult: результат успешно отправлен\n")
			return nil
		} else {
			fmt.Printf("SendResult: задача уже выполнена (Result=%s), результат не отправляется\n", task.Result)
		}
	} else {
		fmt.Printf("SendResult: выражение %s не найдено в ExpressionMap\n", strId)
	}
	return calc.ErrNotFound
}

func getToken(login string) (string, error) {

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name": login,
		"nbf":  now.Unix(),
		"exp":  now.Add(contract.TokenExpiredTimeHours * time.Hour).Unix(),
		"iat":  now.Unix(),
	})

	tokenString, err := token.SignedString([]byte(contract.CalcServerSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func CheckToken(tokenString string) (string, error) {

	tokenFromString, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return []byte(""), errors.New("BAD AUTORIZATION TOKEN")
		}
		return []byte(contract.CalcServerSecret), nil
	})

	if err != nil {
		return "", err
	}

	userLogin := ""
	if claims, ok := tokenFromString.Claims.(jwt.MapClaims); ok {
		if claims != nil {
			userLogin, ok = claims["name"].(string)
			if !ok {
				return "", errors.New("BAD AUTORIZATION TOKEN")
			}
		}
	} else {
		return "", errors.New("BAD AUTORIZATION TOKEN")
	}

	return userLogin, nil
}
