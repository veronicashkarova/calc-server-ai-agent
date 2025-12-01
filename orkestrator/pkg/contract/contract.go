package contract

type Config struct {
	Addr                    string
	TIME_ADDITION_MS        int
	TIME_SUBTRACTION_MS     int
	TIME_MULTIPLICATIONS_MS int
	TIME_DIVISIONS_MS       int
}

type TokenData struct {
	Token string `json:"token"`
}

type RequestData struct {
	Expression string `json:"expression"`
}

type UserLogin struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type ResponseData struct {
	ID string `json:"id"`
}

type ExpressionsData struct {
	Expressions []ExpressionData `json:"expressions"`
}

type ExpressionData struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Result string `json:"result"`
}

type TaskData struct {
	ID            int     `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

type TaskResult struct {
	ID     int     `json:"id"`
	Result float64 `json:"result"`
}

type ExpressionMapData struct {
	User    string
	Data    ExpressionData
	ExpChan chan float64
}

const CalcServerSecret = "calc_server_signature"
const TokenExpiredTimeHours = 24

var (
	InProcess     = "IN PROGRESS"
	Done          = "DONE"
	Undefined     = "UNKNOWN"
	AppConfig     *Config
	ExpressionMap = make(map[string]ExpressionMapData)
	TaskChannel   = make(chan TaskData, 100)
)
