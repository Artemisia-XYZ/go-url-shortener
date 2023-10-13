package handlers

import (
	"strings"
	"url-shortener/domain"
	"url-shortener/usecases"
	"url-shortener/utils/validation"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type shortLinkHandler struct {
	shortLinkUcase domain.ShortLinkUsecase
}

func NewShortLinkHandler(shortLinkUcase domain.ShortLinkUsecase) *shortLinkHandler {
	return &shortLinkHandler{shortLinkUcase}
}

func (h *shortLinkHandler) CreateShortLink(c *fiber.Ctx) error {
	req := &domain.CreateShortLinkRequest{}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "unprocessable entity",
		})
	}

	if req.Destination == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "destination is required",
		})
	} else if !strings.Contains(req.Destination, ".") && !strings.Contains(req.Destination, ":") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "destination invalid",
		})
	} else if !strings.Contains(req.Destination, "://") {
		req.Destination = "https://" + req.Destination
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": errs[0].Message,
		})
	}

	shortLink, err := h.shortLinkUcase.CreateShortLink(req)
	if err != nil {
		if err == usecases.ErrSlashCodeExists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	shortLink.Origin = c.BaseURL() + "/" + shortLink.SlashCode

	return c.Status(fiber.StatusCreated).JSON(shortLink)
}

func (h *shortLinkHandler) Redirect(c *fiber.Ctx) error {
	slash := c.Params("slash")
	dest, err := h.shortLinkUcase.Redirect(slash)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	c.Set("Cache-Control", "max-age=180")
	return c.Redirect(dest, fiber.StatusMovedPermanently)
}
