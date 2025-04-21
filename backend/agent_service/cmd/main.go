package main

import application "github.com/asiafrolova/Multi-user-calculator/agent_service/internal/app"

func main() {
	a := application.New()
	a.RunAgent()
}
