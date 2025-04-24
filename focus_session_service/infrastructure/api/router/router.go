package router

import (
	"focusspot/focussessionservice/infrastructure/api/handler"
	"focusspot/focussessionservice/utils/middleware"
	"focusspot/focussessionservice/utils/token"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all routes for API
func SetupRoutes(app *fiber.App, sessionHandler *handler.FocusSessionHandler, tokenMaker token.Maker) {
	// Middleware
	app.Use(middleware.LoggerMiddleware())

	// API routes
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Protected routes - all routes need authentication
	sessions := v1.Group("/focus-sessions")
	sessions.Use(middleware.AuthMiddleware(tokenMaker))

	// Session management
	sessions.Post("/", sessionHandler.CreateSession)
	sessions.Get("/", sessionHandler.GetUserSessions)
	sessions.Get("/active", sessionHandler.GetActiveSession)
	sessions.Get("/:id", sessionHandler.GetSessionByID)
	sessions.Put("/:id", sessionHandler.UpdateSession)
	sessions.Delete("/:id", sessionHandler.DeleteSession)
	
	// Session status management
	sessions.Post("/:id/start", sessionHandler.StartSession)
	sessions.Post("/:id/end", sessionHandler.EndSession)
	sessions.Post("/:id/cancel", sessionHandler.CancelSession)
	
	// Productivity analytics
	sessions.Get("/analytics/stats", sessionHandler.GetProductivityStats)
	sessions.Get("/analytics/trends", sessionHandler.GetProductivityTrends)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
		})
	})
}
