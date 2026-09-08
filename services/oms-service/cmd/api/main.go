package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure" // Needed for local connections without SSL

	pb "omnistock-ai/proto/inventory/v1"
	"omnistock-ai/services/oms-service/internal/handler/http"
	"omnistock-ai/services/oms-service/internal/repository/postgres"
	"omnistock-ai/services/oms-service/internal/usecase"
)

func main() {
	// 1. Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	dbURL := os.Getenv("OMS_DB_URL")
	serverPort := os.Getenv("OMS_PORT")
	inventoryGrpcAddr := os.Getenv("INVENTORY_GRPC_ADDRESS") // Load the phone number!

	if dbURL == "" || serverPort == "" || inventoryGrpcAddr == "" {
		log.Fatal("Error: Required environment variables are missing in .env")
	}

	// 2. Connect to Database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	fmt.Println("✅ Successfully connected to OMS Database!")

	// 3. Connect to Inventory gRPC Server (Dialing...)
	// We use insecure.NewCredentials() because we are running locally without HTTPS/SSL
	conn, err := grpc.NewClient(inventoryGrpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to inventory gRPC server: %v", err)
	}
	defer conn.Close() // Close connection when the server stops

	// Create the "phone" object (Client)
	inventoryClient := pb.NewInventoryServiceClient(conn)
	fmt.Printf("📞 Connected to Inventory gRPC at %s\n", inventoryGrpcAddr)

	// 4. Dependency Injection (Hand the phone to the Usecase)
	repo := postgres.NewOrderRepository(db)
	uc := usecase.NewOrderUsecase(repo, inventoryClient) // <--- Look here!

	// 5. Setup Router
	router := gin.Default()
	http.NewOrderHandler(router, uc)

	// 6. Start Server
	fmt.Printf("🚀 Starting OMS Service on port %s...\n", serverPort)
	if err := router.Run(":" + serverPort); err != nil {
		log.Fatalf("failed to run HTTP server: %v", err)
	}
}
