package routes

import (
	"time"
	"url-shortener/handlers"
	"url-shortener/middleware"

	"github.com/gofiber/fiber/v2"
)

func NewWebRoutes(r fiber.Router, h *handlers.Factory) {
	r.Get("/:slash", middleware.Limiter(1000, 1*time.Hour), h.ShortLink.Redirect)
}
