package main

import application "github.com/asiafrolova/Final_task/agent_service/internal/app"

func main() {
	a := application.New()
	a.RunAgent()
}
