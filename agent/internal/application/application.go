package application

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/veronicashkarova/agent/pkg/agent"
)

type Config struct {
	COMPUTING_POWER int
	IDLE_DELAY      int
	SERVER_HOST     string
	API_KEY         string
	USE_AI          bool
}

func ConfigFromEnv() *Config {

	config := new(Config)
	power, err := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if err != nil {
		config.COMPUTING_POWER = 3
	} else {
		config.COMPUTING_POWER = power
	}

	idleDelay, err := strconv.Atoi(os.Getenv("IDLE_DELAY"))
	if err != nil {
		config.IDLE_DELAY = 5000
	} else {
		config.IDLE_DELAY = idleDelay
	}

	// Получаем IP адрес сервера из переменной окружения или запрашиваем интерактивно
	serverHost := os.Getenv("SERVER_HOST")
	if serverHost == "" {
		serverHost = requestServerHost()
	}
	config.SERVER_HOST = serverHost

	// Получаем API ключ из переменной окружения
	apiKey := os.Getenv("API_KEY")
	config.API_KEY = apiKey

	// Проверяем, нужно ли использовать AI (если указан API_KEY, то используем AI)
	useAI := os.Getenv("USE_AI")
	config.USE_AI = useAI == "true" || apiKey != ""

	return config
}

func requestServerHost() string {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Введите IP адрес сервера (например, 192.168.0.89): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Ошибка чтения ввода: %v\n", err)
			continue
		}

		host := strings.TrimSpace(input)

		// Проверяем, что это валидный IP адрес
		if host == "" {
			fmt.Println("IP адрес не может быть пустым. Попробуйте снова.")
			continue
		}

		ip := net.ParseIP(host)
		if ip == nil {
			fmt.Println("Неверный формат IP адреса. Попробуйте снова.")
			continue
		}

		return host
	}
}

type Application struct {
	config *Config
}

func New() *Application {
	config := ConfigFromEnv()
	return &Application{
		config: config,
	}
}

func (a *Application) RunAgent() {
	// Всегда запускаем смешанный режим: 2 обычных + 1 AI (если указан API_KEY)
	// Передаем API_KEY (может быть пустым), AI агент запустится только если ключ указан
	log.Printf("Запуск агентов: API_KEY длина = %d, USE_AI = %v\n", len(a.config.API_KEY), a.config.USE_AI)
	agent.RunGrpcAgentAI(a.config.COMPUTING_POWER, a.config.IDLE_DELAY, a.config.SERVER_HOST, a.config.API_KEY)
}
