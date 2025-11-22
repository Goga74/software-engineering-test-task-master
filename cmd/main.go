package main

import (
	"cruder/internal/controller"
	"cruder/internal/handler"
	"cruder/internal/repository"
	"cruder/internal/service"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"
	}

	// Load API key from environment variable
	apiKey := os.Getenv("X_API_KEY")
	if apiKey == "" {
		// Default API key for development/testing
		apiKey = "dev-api-key-12345"
		log.Println("Warning: Using default API key. Set X_API_KEY environment variable for production.")
	}

	dbConn, err := repository.NewPostgresConnection(dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	repositories := repository.NewRepository(dbConn.DB())
	services := service.NewService(repositories)
	controllers := controller.NewController(services)
	r := gin.Default()
	handler.New(r, controllers.Users, apiKey)
	if err := r.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
