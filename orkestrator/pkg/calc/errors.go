package calc

import "errors"

var (
	ErrInvalidExpression = errors.New("неправильное выражение")
	ErrNullDivision      = errors.New("деление на ноль")
	ErrIllegalSign       = errors.New("неправильный символ")
	ErrMissingBracket    = errors.New("не хватает скобки")
	ErrEmptyExpression   = errors.New("пустое выражение")
	ErrNotFound          = errors.New("не найдено выражение")
	ErrNotTask           = errors.New("нет доступных задач")
)
