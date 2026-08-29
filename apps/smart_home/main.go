package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"smarthome/db"
	"smarthome/handlers"
	"smarthome/services"
	"smarthome/timeseries"

	"github.com/gin-gonic/gin"
)

func main() {
	// Set up database connection
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/smarthome")
	database, err := db.New(dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer database.Close()

	log.Println("Connected to database successfully")

	clickhouseClient, err := timeseries.New(
		getEnv("CLICKHOUSE_HOST", "clickhouse"),
		getEnv("CLICKHOUSE_PORT", "9000"),
		getEnv("CLICKHOUSE_DATABASE", "default"),
		getEnv("CLICKHOUSE_USERNAME", "smarthome"),
		getEnv("CLICKHOUSE_PASSWORD", "clickhouse"),
	)
	if err != nil {
		log.Fatalf("Unable to create timeseries client: %v\n", err)
	}

	err = clickhouseClient.Connect()
	if err != nil {
		log.Fatalf("Unable to connect to timeseries client: %v\n", err)
	}

	defer clickhouseClient.Close()

	// Initialize temperature service
	temperatureAPIURL := getEnv("TEMPERATURE_API_URL", "http://temperature-api:8081")
	temperatureService := services.NewTemperatureService(temperatureAPIURL)
	log.Printf("Temperature service initialized with API URL: %s\n", temperatureAPIURL)

	commandServiceAddr := getEnv("COMMAND_SERVICE_ADDR", "command-service:50051")
	commandService, err := services.NewCommandService(commandServiceAddr)
	if err != nil {
		log.Fatalf("Unable to create command service client: %v\n", err)
	}
	log.Printf("Command service client initialized with address: %s\n", commandServiceAddr)

	metricsService := services.NewMetricsService(clickhouseClient)

	// Initialize router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// API routes
	apiRoutes := router.Group("/api/v1")

	// Register sensor routes
	sensorHandler := handlers.NewSensorHandler(database, temperatureService, metricsService)
	sensorHandler.RegisterRoutes(apiRoutes)

	commandHandler := handlers.NewCommandHandler(commandService)
	commandHandler.RegisterRoutes(apiRoutes)

	// Start server
	srv := &http.Server{
		Addr:    getEnv("PORT", ":8080"),
		Handler: router,
	}

	// Start the server in a goroutine
	go func() {
		log.Printf("Server starting on %s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v\n", err)
	}

	log.Println("Server exited properly")
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
