package main

import (
	"database/sql"
	"fmt"
	"log"
	"net" // Required for gRPC network listener
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc" // Required for gRPC server

	pb "omnistock-ai/proto/inventory/v1"
	grpchandler "omnistock-ai/services/inventory-service/internal/handler/grpc"
	"omnistock-ai/services/inventory-service/internal/handler/http"
	"omnistock-ai/services/inventory-service/internal/repository/postgres"
	"omnistock-ai/services/inventory-service/internal/usecase"
)

func main() {
	// 1. Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, falling back to system environment variables")
	}

	dbURL := os.Getenv("INVENTORY_DB_URL")
	httpPort := os.Getenv("INVENTORY_PORT")
	grpcPort := os.Getenv("INVENTORY_GRPC_PORT")

	if dbURL == "" || httpPort == "" || grpcPort == "" {
		log.Fatal("Error: Required environment variables are missing")
	}

	// 2. Setup Database Connection
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	fmt.Println("✅ Successfully connected to PostgreSQL!")

	// 3. Dependency Injection
	repo := postgres.NewProductRepository(db)
	uc := usecase.NewProductUsecase(repo)

	// 4. Setup and Start gRPC Server in a background Goroutine
	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("failed to listen on gRPC port: %v", err)
		}

		grpcServer := grpc.NewServer()
		inventoryGrpcHandler := grpchandler.NewInventoryGrpcHandler(uc)
		pb.RegisterInventoryServiceServer(grpcServer, inventoryGrpcHandler)

		fmt.Printf("📞 Starting Inventory gRPC Server on port %s...\n", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// 5. Setup and Start HTTP Server (Blocks the main thread)
	router := gin.Default()
	http.NewProductHandler(router, uc)

	fmt.Printf("🚀 Starting Inventory HTTP Service on port %s...\n", httpPort)
	if err := router.Run(":" + httpPort); err != nil {
		log.Fatalf("failed to run HTTP server: %v", err)
	}
}
