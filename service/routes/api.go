package routes

import (
	"time"
	"url-shortener/handlers"
	"url-shortener/middleware"

	"github.com/gofiber/fiber/v2"
)

func NewAPIRoutes(r fiber.Router, h *handlers.Factory) {
	r.Post("/links", middleware.Limiter(150, 1*time.Hour), h.ShortLink.CreateShortLink)
}
