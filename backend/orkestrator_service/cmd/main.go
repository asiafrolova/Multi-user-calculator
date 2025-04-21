package main

import (
	application "github.com/asiafrolova/Final_task/orkestrator_service/internal/app"
)

func main() {
	app := application.New()
	app.RunServer()

}
