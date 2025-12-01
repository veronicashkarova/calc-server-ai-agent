package application

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"

	"github.com/veronicashkarova/server-for-calc/pkg/orkestrator"
	pb "github.com/veronicashkarova/server-for-calc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Server struct {
	pb.CalculatorServiceServer // сервис из сгенерированного пакета
}

func NewServer() *Server {
	return &Server{}
}

type CalculatorServiceServer interface {
	GetTask(context.Context, *pb.EmptyRequest) (*pb.Task, error)
	GetResult(context.Context, *pb.TaskResult) *pb.EmptyResponse
	mustEmbedUnimplementedGeometryServiceServer()
}

func (s *Server) GetTask(
	ctx context.Context,
	req *pb.EmptyRequest,
) (*pb.Task, error) {
	task, err := orkestrator.GetTaskData()
	fmt.Printf("GetTask: получена задача из канала: ID=%d, Arg1=%f, Arg2=%f, Operation=%s, err=%v\n", task.ID, task.Arg1, task.Arg2, task.Operation, err)
	if err != nil {
		// Возвращаем ошибку, если задач нет, вместо пустой задачи
		fmt.Printf("GetTask: ошибка получения задачи: %v\n", err)
		return nil, err
	}
	// Проверяем, что задача валидна (ID не равен 0)
	if task.ID == 0 {
		fmt.Printf("GetTask: получена невалидная задача с ID=0\n")
		return nil, fmt.Errorf("invalid task: task ID is zero")
	}
	fmt.Printf("GetTask: возвращаем задачу агенту: ID=%d\n", task.ID)
	return &pb.Task{
		Id:            int32(task.ID),
		Arg1:          float32(task.Arg1),
		Arg2:          float32(task.Arg2),
		Operation:     task.Operation,
		OperationTime: int32(task.OperationTime),
	}, nil
}

func (s *Server) GetResult(
	ctx context.Context,
	taskResult *pb.TaskResult,
) (*pb.EmptyResponse, error) {
	fmt.Printf("GetResult: получен результат от агента: ID=%d, Result=%f\n", taskResult.Id, taskResult.Result)
	resp := &pb.EmptyResponse{}
	var resultErr = orkestrator.SendResult(int(taskResult.Id), float64(taskResult.Result))
	if resultErr != nil {
		fmt.Printf("GetResult: ошибка отправки результата: %v\n", resultErr)
		return resp, resultErr
	}
	fmt.Printf("GetResult: результат успешно обработан\n")
	return resp, nil
}

func StartGrpcServer() {
	go func() {
		host := "0.0.0.0" //localhost
		port := "5000"

		addr := fmt.Sprintf("%s:%s", host, port)
		lis, err := net.Listen("tcp", addr) // будем ждать запросы по этому адресу

		if err != nil {
			fmt.Println("error starting tcp listener: ", err)
			os.Exit(1)
		}

		// Загружаем TLS сертификаты
		cert, err := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")
		if err != nil {
			fmt.Println("error loading TLS certificates: ", err)
			fmt.Println("Please run: cd orkestrator && bash scripts/generate_certs.sh")
			os.Exit(1)
		}

		// Создаем TLS конфигурацию
		// Используем минимальные настройки для совместимости
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		// Создаем TLS credentials
		creds := credentials.NewTLS(tlsConfig)

		fmt.Println("tcp listener started at port: ", port)
		// создадим сервер grpc с TLS
		grpcServer := grpc.NewServer(grpc.Creds(creds))
		// объект структуры, которая содержит реализацию
		// серверной части GeometryService
		calcServiceServer := NewServer()
		// зарегистрируем нашу реализацию сервера
		pb.RegisterCalculatorServiceServer(grpcServer, calcServiceServer)
		// запустим grpc сервер
		if err := grpcServer.Serve(lis); err != nil {
			fmt.Println("error serving grpc: ", err)
			os.Exit(1)
		}
	}()
}
