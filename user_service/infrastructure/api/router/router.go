package router

import (
	"focusspot/userservice/infrastructure/api/handler"
	"focusspot/userservice/utils/middleware"
	"focusspot/userservice/utils/token"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all routes for API
func SetupRoutes(app *fiber.App, userHandler *handler.UserHandler, tokenMaker token.Maker) {
	// Middleware
	app.Use(middleware.LoggerMiddleware())

	// API routes
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Public routes
	auth := v1.Group("/auth")
	auth.Post("/register", userHandler.Register)
	auth.Post("/login", userHandler.Login)

	// Protected routes
	user := v1.Group("/users")
	user.Use(middleware.AuthMiddleware(tokenMaker))
	user.Get("/me", userHandler.GetProfile)
	user.Put("/me", userHandler.UpdateProfile)
	user.Put("/me/preferences", userHandler.UpdatePreferences)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
		})
	})
}
