package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"order-service/handler"
	"order-service/repository"
	"order-service/service"
)

func main() {
	// Load .env file — silently ignore if not found (e.g. inside Docker)
	_ = godotenv.Load()

	// Read configuration from environment variables
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	dbName := os.Getenv("MONGO_DATABASE")
	if dbName == "" {
		dbName = "order_db"
	}

	productServiceURL := os.Getenv("PRODUCT_SERVICE_URL")
	if productServiceURL == "" {
		productServiceURL = "http://localhost:8001"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}

	// Connect to MongoDB
	connectCtx, connectCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer connectCancel()

	client, err := mongo.Connect(connectCtx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("failed to connect to mongodb: %v", err)
	}

	if err := client.Ping(connectCtx, nil); err != nil {
		log.Fatalf("failed to ping mongodb: %v", err)
	}
	log.Println("connected to mongodb")

	db := client.Database(dbName)

	// Wire dependencies: repository → service → handler
	orderRepo := repository.NewOrderRepository(db)
	orderSvc := service.NewOrderService(orderRepo, productServiceURL)
	orderHandler := handler.NewOrderHandler(orderSvc)

	// Setup Fiber app and register routes
	app := fiber.New()
	orderHandler.RegisterRoutes(app)

	// Start server in a goroutine
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	log.Printf("order-service running on port %s", port)

	// Graceful shutdown on SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")

	if err := app.Shutdown(); err != nil {
		log.Printf("error shutting down server: %v", err)
	}

	disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer disconnectCancel()

	if err := client.Disconnect(disconnectCtx); err != nil {
		log.Printf("error disconnecting from mongodb: %v", err)
	}

	log.Println("server stopped")
}
