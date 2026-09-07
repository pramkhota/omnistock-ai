package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // Important: anonymous import for postgres driver

	"omnistock-ai/services/inventory-service/internal/handler/http"
	"omnistock-ai/services/inventory-service/internal/repository/postgres"
	"omnistock-ai/services/inventory-service/internal/usecase"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, falling back to system environment variables")
	}

	dbURL := os.Getenv("INVENTORY_DB_URL")
	serverPort := os.Getenv("INVENTORY_PORT")

	if dbURL == "" || serverPort == "" {
		log.Fatal("Error: INVENTORY_DB_URL or INVENTORY_PORT is not set in .env")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	fmt.Println("Successfully connected to PostgreSQL!")

	// 2. Dependency Injection (Wiring layers together)
	repo := postgres.NewProductRepository(db)
	uc := usecase.NewProductUsecase(repo)

	// 3. Setup Router and Handler
	router := gin.Default()
	http.NewProductHandler(router, uc)

	// 4. Start the server on port 8001
	fmt.Println("Starting Inventory Service on port 8001...")
	if err := router.Run(":8001"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
