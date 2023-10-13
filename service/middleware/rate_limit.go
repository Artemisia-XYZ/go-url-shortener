package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func Limiter(requests int, reset time.Duration) fiber.Handler {
	limiter := limiter.New(limiter.Config{
		Max:        requests,
		Expiration: reset,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"message": "rate limit exceeded",
			})
		},
	})

	return func(c *fiber.Ctx) error {
		return limiter(c)
	}
}
