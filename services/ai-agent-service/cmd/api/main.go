package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"omnistock-ai/services/ai-agent-service/internal/handler/http"
	"omnistock-ai/services/ai-agent-service/internal/usecase"
)

func main() {
	// 1. Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	serverPort := os.Getenv("AI_AGENT_PORT")
	if serverPort == "" {
		log.Fatal("Error: AI_AGENT_PORT is not set in .env")
	}

	// 2. Initialize the AI Usecase (Connects to Google Gemini)
	agentUC, err := usecase.NewAgentUsecase()
	if err != nil {
		log.Fatalf("failed to initialize AI Agent: %v", err)
	}
	fmt.Println("🧠 Successfully connected to Google Gemini AI!")

	// 3. Setup Router and Handlers
	router := gin.Default()
	http.NewAgentHandler(router, agentUC)

	// 4. Start HTTP Server
	fmt.Printf("🚀 Starting AI Agent Service on port %s...\n", serverPort)
	if err := router.Run(":" + serverPort); err != nil {
		log.Fatalf("failed to run HTTP server: %v", err)
	}
}
