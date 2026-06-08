package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file — silently ignore if not found (e.g. inside Docker)
	_ = godotenv.Load()

	productServiceURL := os.Getenv("PRODUCT_SERVICE_URL")
	if productServiceURL == "" {
		productServiceURL = "http://localhost:8001"
	}

	orderServiceURL := os.Getenv("ORDER_SERVICE_URL")
	if orderServiceURL == "" {
		orderServiceURL = "http://localhost:8002"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	app := fiber.New()

	// CORS middleware — allows frontend to fetch without errors
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type",
	}))

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Proxy /products and /products/* → product-service
	// Note: /internal/* is NOT proxied — only /products paths are exposed
	app.All("/products/*", createProxyHandler(productServiceURL, "product service"))
	app.All("/products", createProxyHandler(productServiceURL, "product service"))

	// Proxy /orders and /orders/* → order-service
	app.All("/orders/*", createProxyHandler(orderServiceURL, "order service"))
	app.All("/orders", createProxyHandler(orderServiceURL, "order service"))

	// Start server in a goroutine
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	log.Printf("gateway running on port %s", port)

	// Graceful shutdown on SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down gateway...")
	if err := app.Shutdown(); err != nil {
		log.Printf("error shutting down: %v", err)
	}
	log.Println("gateway stopped")
}

// createProxyHandler creates a Fiber handler that proxies requests to the target service.
// On upstream failure, returns 502 Bad Gateway with a descriptive error message.
func createProxyHandler(targetURL string, serviceName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		url := targetURL + c.OriginalURL()
		if err := proxy.Do(c, url); err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": serviceName + " unavailable"})
		}
		// Remove upstream server header
		c.Response().Header.Del(fiber.HeaderServer)
		return nil
	}
}
