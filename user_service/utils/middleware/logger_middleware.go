package middleware

import (
	"github.com/gofiber/fiber/v2"

	"github.com/gofiber/fiber/v2/middleware/logger"
)

// LoggerMiddle returns configured middleware logger
func LoggerMiddleware() fiber.Handler {
	return logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${latency} ${method} ${path}\n",
		TimeFormat: "002-Jan-2006 15:04:05",
		TimeZone:   "Local",
	})
}
