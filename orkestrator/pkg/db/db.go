package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"github.com/veronicashkarova/server-for-calc/pkg/contract"
)

type (
	User struct {
		ID       int64
		Login    string
		Password string
	}

	Expression struct {
		ID         int64
		Expression string
		UserID     int64
		Status     string
		Result     string
	}
)

var ctx = context.TODO()
var db *sql.DB = nil

func createTables(ctx context.Context, db *sql.DB) error {
	const (
		usersTable = `
	CREATE TABLE IF NOT EXISTS users(
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		login TEXT,
		password TEXT
	);`

		expressionsTable = `
	CREATE TABLE IF NOT EXISTS expressions(
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		expression TEXT NOT NULL,
		user_id INTEGER NOT NULL,
		status TEXT NOT NULL,
		result TEXT NOT NULL,
	
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

func InsertUser(user *contract.UserLogin) (int64, error) {

	if CheckUser(user) {
		return 0, errors.New("USER ALREADY REGISTERED")
	}

	var q = `
	INSERT INTO users (login, password) values ($1, $2)
	`
	result, err := db.ExecContext(ctx, q, user.Login, user.Password)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func CheckUser(user *contract.UserLogin) bool {
	var q = `
	SELECT id FROM users WHERE login = $1 LIMIT 1
	`

	var userID int64
	row := db.QueryRow(q, user.Login)

	err := row.Scan(&userID)

	if err == sql.ErrNoRows {
		return false
	} else {
		return true
	}
}

func InsertExpression(expression *Expression) (int64, error) {
	var q = `
	INSERT INTO expressions (expression, user_id, status, result) values ($1, $2, $3, $4)
	`

	result, err := db.ExecContext(ctx, q, expression.Expression, expression.UserID, expression.Status, expression.Result)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (u User) Print() string {
	id := strconv.FormatInt(u.ID, 10)
	return "ID: " + id + " Login: " + u.Login + " Password: " + u.Password
}

func (e Expression) Print() string {
	id := strconv.FormatInt(e.ID, 10)
	userID := strconv.FormatInt(e.UserID, 10)
	return "ID: " + id + " Expression" + e.Expression + " UserID:" + userID
}

func selectUsers(ctx context.Context, db *sql.DB) ([]User, error) {
	var users []User
	var q = "SELECT id, login, password FROM users"
	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		u := User{}
		err := rows.Scan(&u.ID, &u.Login, &u.Password)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

func SelectExpressionsForUserId(userId int64) ([]Expression, error) {
	var expressions []Expression
	var q = "SELECT id, expression, user_id, status, result FROM expressions WHERE user_id = $1"

	rows, err := db.QueryContext(ctx, q, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		e := Expression{}
		err := rows.Scan(&e.ID, &e.Expression, &e.UserID, &e.Status, &e.Result)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, e)
	}

	return expressions, nil
}

func SelectExpressionForId(id int64) (Expression, error) {
	var u Expression
	var q = "SELECT id, expression, user_id, status, result FROM expressions WHERE id = $1"

	err := db.QueryRowContext(ctx, q, id).Scan(&u.ID, &u.Expression, &u.UserID, &u.Status, &u.Result)

	if err != nil {
		return u, err
	}

	return u, nil
}

func UpdateExpressionStatusResult(id int64, newStatus string, newResult string) error {
	var q = "UPDATE expressions SET status = $1, result = $2 WHERE id = $3"

	if ctx.Err() != nil {
		return fmt.Errorf("контекст истек: %w", ctx.Err())
	}

	_, err := db.ExecContext(ctx, q, newStatus, newResult, id)
	if err != nil {
		return fmt.Errorf("ошибка выполнения запроса: %w", err)
	}

	return nil
}

func selectExpressions(ctx context.Context, db *sql.DB) ([]Expression, error) {
	var expressions []Expression
	var q = "SELECT id, expression, user_id FROM expressions"

	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		e := Expression{}
		err := rows.Scan(&e.ID, &e.Expression, &e.UserID)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, e)
	}

	return expressions, nil
}

func selectUserByID(ctx context.Context, db *sql.DB, id int64) (User, error) {
	u := User{}
	var q = "SELECT id, login, password FROM users WHERE id = $1"
	err := db.QueryRowContext(ctx, q, id).Scan(&u.ID, &u.Login, &u.Password)
	if err != nil {
		return u, err
	}

	return u, nil
}

func SelectIdForUser(userLogin string) (int64, error) {
	u := User{}
	var q = "SELECT id, login, password FROM users WHERE login = $1"
	err := db.QueryRowContext(ctx, q, userLogin).Scan(&u.ID, &u.Login, &u.Password)
	if err != nil {
		return 0, err
	}

	return u.ID, nil
}

func CreateDb() {
	var err error = nil

	db, err = sql.Open("sqlite3", "store.db")
	if err != nil {
		panic(err)
	}

	fmt.Println("DB opened")

	err = db.PingContext(ctx)
	if err != nil {
		panic(err)
	}

	if err = createTables(ctx, db); err != nil {
		panic(err)
	}
}
