package main

import (
	"context"
	"fmt"
	usecase "focusspot/focussessionservice/application/usecases"
	"focusspot/focussessionservice/config"
	"focusspot/focussessionservice/infrastructure/api/handler"
	"focusspot/focussessionservice/infrastructure/api/router"
	"focusspot/focussessionservice/infrastructure/persistence/mongodb"
	"focusspot/focussessionservice/utils/token"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Load configs
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to MongoDB
	client, db, err := mongodb.NewMongoDBConnection(&cfg.MongoDB)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongodb.CloseMongoDBConnection(client)

	// Create JWT token maker
	tokenMaker, err := token.NewJWTMaker(cfg.JWT.SecretKey)
	if err != nil {
		log.Fatalf("Failed to create token maker: %v", err)
	}

	// Setup repositories
	sessionRepo := mongodb.NewMongoFocusSessionRepository(db)

	// Setup usecases
	sessionUseCase := usecase.NewFocusSessionUseCase(sessionRepo)

	// Setup handlers
	sessionHandler := handler.NewFocusSessionHandler(sessionUseCase)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Setup routes
	router.SetupRoutes(app, sessionHandler, tokenMaker)

	// Start server in a goroutine
	go func() {
		if err := app.Listen(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Shutdown server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
