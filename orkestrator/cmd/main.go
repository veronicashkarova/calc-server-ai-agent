package main

import (
	"github.com/veronicashkarova/server-for-calc/internal/application"
)

func main() {
	app := application.New()
	app.CreareDataBase()
	app.RunServer()
	
}