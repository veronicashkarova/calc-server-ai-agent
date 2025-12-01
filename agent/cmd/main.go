package main

import (
	"github.com/veronicashkarova/agent/internal/application"
)

func main() {
	app := application.New()
	app.RunAgent()
}
