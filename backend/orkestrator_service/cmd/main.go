package main

import application "github.com/asiafrolova/Multi-user-calculator/orkestrator_service/internal/app"

func main() {
	app := application.New()
	app.RunServer()

}
