package middleware

import (
	"github.com/gofiber/fiber/v2"
	"ai2sim/backend/internal/llm"
)

// ConcurrencyLimiter returns a Fiber middleware that rejects requests when the
// LLM client is already at maximum concurrent capacity.
// Only applied to /api/v1/simulate — other endpoints always pass through.
func ConcurrencyLimiter(client *llm.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only gate simulation streaming endpoint
		if c.Path() == "/api/v1/simulate" && client.IsAtCapacity() {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":   "Server at capacity — too many concurrent simulations",
				"code":    "CAPACITY_EXCEEDED",
				"message": "Please wait a moment and try again.",
			})
		}
		return c.Next()
	}
}
