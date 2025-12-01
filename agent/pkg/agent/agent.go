package agent

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	pb "github.com/veronicashkarova/agent/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Task struct {
	ID            int     `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

type Result struct {
	ID     int     `json:"id"`
	Result float64 `json:"result"`
}

func RunGrpcAgent(power int, delay int, host string) {
	fmt.Printf("start agent, connecting to server at %s\n", host)
	runGrpcAgentInternal(power, delay, host, "")
}

func RunGrpcAgentAI(power int, delay int, host string, apiKey string) {
	fmt.Printf("start agent with AI, connecting to server at %s\n", host)
	runGrpcAgentInternal(power, delay, host, apiKey)
}

func runGrpcAgentInternal(power int, delay int, host string, apiKey string) {
	fmt.Printf("start agent, connecting to server at %s\n", host)

	port := "5000"

	var wg sync.WaitGroup
	
	// Запускаем 2 обычных агента, каждый со своим соединением
	for i := 1; i <= 2; i++ {
		wg.Add(1)
		go func(agentNum int) {
			defer wg.Done()
			log.Printf("Запуск обычного агента #%d\n", agentNum)
			// Создаем отдельное соединение для каждого агента
			agentConn, err := createAgentConnection(host, port)
			if err != nil {
				log.Printf("Агент #%d: ошибка создания соединения: %v\n", agentNum, err)
				return
			}
			defer agentConn.Close()
			agentClient := pb.NewCalculatorServiceClient(agentConn)
			startGrpcAgent(agentClient, delay)
		}(i)
		// Задержка в 1 секунду перед запуском следующего агента
		if i < 2 {
			time.Sleep(1 * time.Second)
		}
	}
	
	// Запускаем 1 AI агента (всегда, независимо от наличия API ключа)
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("Запуск AI агента\n")
		// Создаем отдельное соединение для AI агента
		aiConn, err := createAgentConnection(host, port)
		if err != nil {
			log.Printf("AI Агент: ошибка создания соединения: %v\n", err)
			return
		}
		defer aiConn.Close()
		aiClient := pb.NewCalculatorServiceClient(aiConn)
		startGrpcAgentAI(aiClient, delay, apiKey)
	}()
	time.Sleep(1 * time.Second)

	wg.Wait()
}

// createAgentConnection создает новое соединение для агента
func createAgentConnection(host string, port string) (*grpc.ClientConn, error) {
	addr := fmt.Sprintf("%s:%s", host, port)
	
	// Загружаем сертификат сервера
	certPool := x509.NewCertPool()
	certFile, err := os.ReadFile("certs/server.crt")
	if err != nil {
		return nil, fmt.Errorf("error reading server certificate: %v", err)
	}

	if !certPool.AppendCertsFromPEM(certFile) {
		return nil, fmt.Errorf("failed to append server certificate")
	}

	// Создаем TLS конфигурацию
	tlsConfig := &tls.Config{
		RootCAs:            certPool,
		ServerName:         host,
		InsecureSkipVerify: false,
	}

	// Создаем TLS credentials
	creds := credentials.NewTLS(tlsConfig)

	// Создаем новое соединение
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("could not connect to grpc server at %s: %v", addr, err)
	}

	return conn, nil
}

func startGrpcAgent(client pb.CalculatorServiceClient, delay int) {
	ctx := context.TODO()

	for {
		log.Printf("Агент: запрос задачи у сервера...")
		req, err := client.GetTask(ctx, &pb.EmptyRequest{})

		if err != nil {
			log.Printf("Агент: ошибка получения задачи: %v. Повторная попытка через %d секунд...", err, delay/1000)
			Delay(delay)
			continue
		}

		log.Printf("Агент: получена задача от сервера: ID=%d, Arg1=%f, Arg2=%f, Operation=%s, OperationTime=%d",
			req.Id, req.Arg1, req.Arg2, req.Operation, req.OperationTime)

		// Проверяем, что задача валидна (ID не равен 0)
		if req.Id == 0 {
			log.Printf("Агент: получена невалидная задача с ID=0, пропускаем")
			continue
		}

		task := Task{
			ID:            int(req.Id),
			Arg1:          float64(req.Arg1),
			Arg2:          float64(req.Arg2),
			Operation:     req.Operation,
			OperationTime: int(req.OperationTime),
		}

		operationTimer := time.NewTimer(time.Duration(task.OperationTime * int(time.Millisecond)))
		<-operationTimer.C

		result, err := executeTask(task)
		if err != nil {
			log.Printf("Ошибка выполнения задачи")
			continue
		}

		_, err = client.GetResult(ctx, &pb.TaskResult{
			Id:     int32(result.ID),
			Result: float32(result.Result),
		})

		if err != nil {
			log.Printf("Ошибка отправки результата задачи")
			Delay(delay)
			continue
		}

		fmt.Printf("Задача %d выполнена успешно. Результат: %f\n", task.ID, result.Result)
	}
}

func startGrpcAgentAI(client pb.CalculatorServiceClient, delay int, apiKey string) {
	ctx := context.TODO()

	for {
		log.Printf("AI Агент: запрос задачи у сервера...")
		req, err := client.GetTask(ctx, &pb.EmptyRequest{})

		if err != nil {
			log.Printf("AI Агент: ошибка получения задачи: %v. Повторная попытка через %d секунд...", err, delay/1000)
			Delay(delay)
			continue
		}

		log.Printf("AI Агент: получена задача от сервера: ID=%d, Arg1=%f, Arg2=%f, Operation=%s, OperationTime=%d",
			req.Id, req.Arg1, req.Arg2, req.Operation, req.OperationTime)

		// Проверяем, что задача валидна (ID не равен 0)
		if req.Id == 0 {
			log.Printf("AI Агент: получена невалидная задача с ID=0, пропускаем")
			continue
		}

		task := Task{
			ID:            int(req.Id),
			Arg1:          float64(req.Arg1),
			Arg2:          float64(req.Arg2),
			Operation:     req.Operation,
			OperationTime: int(req.OperationTime),
		}

		operationTimer := time.NewTimer(time.Duration(task.OperationTime * int(time.Millisecond)))
		<-operationTimer.C

		result, err := executeTaskAI(task, apiKey)
		if err != nil {
			log.Printf("AI Агент: ошибка выполнения задачи через API: %v", err)
			continue
		}

		_, err = client.GetResult(ctx, &pb.TaskResult{
			Id:     int32(result.ID),
			Result: float32(result.Result),
		})

		if err != nil {
			log.Printf("AI Агент: ошибка отправки результата задачи")
			Delay(delay)
			continue
		}

		fmt.Printf("AI Агент: задача %d выполнена успешно. Результат: %f\n", task.ID, result.Result)
	}
}

func Delay(delay int) {
	delayTimer := time.NewTimer(time.Duration(delay * int(time.Millisecond)))
	<-delayTimer.C
}

func executeTask(task Task) (Result, error) {
	var result float64
	switch task.Operation {
	case "+":
		result = task.Arg1 + task.Arg2
	case "-":
		result = task.Arg1 - task.Arg2
	case "*":
		result = task.Arg1 * task.Arg2
	case "/":
		result = task.Arg1 / task.Arg2
	default:
		return Result{}, fmt.Errorf("неизвестная операция: %s", task.Operation)
	}

	return Result{ID: task.ID, Result: result}, nil
}

// Структуры для работы с API нейросети
type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

// executeTaskAI выполняет задачу через API нейросети
func executeTaskAI(task Task, apiKey string) (Result, error) {
	// Проверяем наличие API ключа
	if apiKey == "" {
		return Result{}, fmt.Errorf("API_KEY не указан в переменных окружения")
	}
	
	// Формируем запрос к нейросети
	taskDescription := fmt.Sprintf("Реши математическую задачу: %.2f %s %.2f. Верни только число-результат без дополнительных объяснений.",
		task.Arg1, task.Operation, task.Arg2)

	requestBody := ChatCompletionRequest{
		Model: "anthropic/claude-sonnet-4-20250514",
		Messages: []Message{
			{
				Role:    "user",
				Content: taskDescription,
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return Result{}, fmt.Errorf("ошибка сериализации запроса: %v", err)
	}

	// Создаем HTTP запрос
	req, err := http.NewRequest("POST", "https://openai.api.proxyapi.ru/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return Result{}, fmt.Errorf("ошибка создания запроса: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Выполняем запрос
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("ошибка API: статус %d, тело: %s", resp.StatusCode, string(body))
	}

	// Парсим ответ
	var apiResponse ChatCompletionResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return Result{}, fmt.Errorf("ошибка парсинга ответа: %v", err)
	}

	if len(apiResponse.Choices) == 0 {
		return Result{}, fmt.Errorf("пустой ответ от API")
	}

	// Извлекаем результат из ответа нейросети
	resultStr := strings.TrimSpace(apiResponse.Choices[0].Message.Content)
	
	// Удаляем возможные символы форматирования (точки, запятые в конце и т.д.)
	resultStr = strings.Trim(resultStr, ".,!?;: \n\t\r")

	// Парсим число из ответа
	result, err := strconv.ParseFloat(resultStr, 64)
	if err != nil {
		return Result{}, fmt.Errorf("ошибка парсинга результата '%s': %v", string(resultStr), err)
	}

	return Result{ID: task.ID, Result: result}, nil
}
